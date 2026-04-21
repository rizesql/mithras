package jws

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/idkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

const (
	keySigningDuration   = 24 * time.Hour
	keyGracePeriod       = 30 * time.Minute
	keySyncInterval      = 5 * time.Minute
	keyRotationLookahead = keyGracePeriod // wake up when this close to rotates_at
	keyRotationBackoff   = 5 * time.Minute
)

type DBStore struct {
	db         *db.Database
	enc        *encryptor
	codec      KeyCodec
	mu         sync.RWMutex
	rotMu      sync.Mutex
	activeKeys []Key
	signingKey Key
}

func NewDBStore(
	ctx context.Context,
	db *db.Database,
	kek []byte,
	codec KeyCodec,
) (*DBStore, error) {
	enc, err := newEncryptor(kek)
	if err != nil {
		return nil, err
	}

	s := &DBStore{db: db, enc: enc, codec: codec}
	if err := s.bootstrap(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *DBStore) bootstrap(ctx context.Context) (err error) {
	ctx, span := telemetry.Start(ctx, "jws.bootstrap")
	defer telemetry.End(span, &err)

	if err = s.sync(ctx); err != nil {
		return fmt.Errorf("bootstrap sync failed: %w", err)
	}

	s.mu.RLock()
	hasSigning := s.signingKey.Signer != nil && s.signingKey.RotatesAt.After(time.Now())
	s.mu.RUnlock()

	if !hasSigning {
		s.rotMu.Lock()
		err = s.rotate(ctx)
		s.rotMu.Unlock()

		if err != nil {
			return fmt.Errorf("bootstrap rotate failed: %w", err)
		}
	}

	return nil
}

func (s *DBStore) sync(ctx context.Context) (err error) {
	ctx, span := telemetry.Start(ctx, "jws.sync")
	defer telemetry.End(span, &err)

	now := time.Now()

	if err = db.Query.PruneJWS(ctx, s.db, now); err != nil {
		_ = telemetry.Err(ctx, fmt.Errorf("prune failed: %w", err))
	}

	rows, err := db.Query.GetActiveJWSKeys(ctx, s.db, now)
	if err != nil {
		return fmt.Errorf("fetch active keys: %w", err)
	}

	active := make([]Key, 0, len(rows))
	for _, row := range rows {
		raw, err := s.enc.Decrypt(row.Data)
		if err != nil {
			_ = telemetry.Err(ctx, fmt.Errorf("decrypt failed for key %s: %w", row.ID, err))
			continue
		}

		signer, err := s.codec.Decode(raw)
		if err != nil {
			_ = telemetry.Err(ctx, fmt.Errorf("decode failed for key %s: %w", row.ID, err))
			continue
		}

		active = append(active, Key{
			ID:        row.ID,
			Signer:    signer,
			Alg:       s.codec.Alg(),
			RotatesAt: row.RotatesAt,
			ExpiresAt: row.ExpiresAt,
		})
	}

	var signing Key

	for _, k := range active {
		if k.RotatesAt.After(now) {
			if signing.Signer == nil || k.RotatesAt.After(signing.RotatesAt) {
				signing = k
			}
		}
	}

	telemetry.Attr(ctx, attribute.Int("jws.active_keys_count", len(active)))

	if signing.Signer != nil {
		telemetry.Attr(ctx, attribute.String("jws.signing_key_id", signing.ID.String()))
	}

	s.mu.Lock()
	s.activeKeys = active
	s.signingKey = signing
	s.mu.Unlock()

	return nil
}

func (s *DBStore) rotate(ctx context.Context) (err error) {
	ctx, span := telemetry.Start(ctx, "jws.rotate")
	defer telemetry.End(span, &err)

	s.mu.RLock()
	needsRotation := s.signingKey.Signer == nil || time.Until(s.signingKey.RotatesAt) <= keyRotationLookahead
	s.mu.RUnlock()

	if !needsRotation {
		telemetry.Attr(ctx, attribute.Bool("jws.rotation_skipped", true))
		return nil
	}

	_, raw, err := s.codec.New()
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	encrypted, err := s.enc.Encrypt(raw)
	if err != nil {
		return fmt.Errorf("encrypt key: %w", err)
	}

	now := time.Now()
	rotatesAt := now.Add(keySigningDuration)
	expiresAt := rotatesAt.Add(keyGracePeriod)

	keyID := idkit.NewKeyID()
	telemetry.Attr(ctx, attribute.String("jws.new_key_id", keyID.String()))

	if err := db.Query.InsertJWSKey(ctx, s.db, db.InsertJWSKeyParams{
		ID:        keyID,
		Data:      encrypted,
		RotatesAt: rotatesAt,
		ExpiresAt: expiresAt,
	}); err != nil {
		return fmt.Errorf("insert key: %w", err)
	}

	// Best-effort sync. If the insert succeeded but sync fails, the background
	// runSync loop will eventually pick it up. This prevents duplicated keys.
	if syncErr := s.sync(ctx); syncErr != nil {
		_ = telemetry.Err(ctx, fmt.Errorf("post-rotation sync failed: %w", syncErr))
	}

	return nil
}

// Sync fires immediately then every keySyncInterval.
// The sync-then-wait pattern means a slow sync naturally delays the next
// tick — concurrent syncs are avoided without additional locking.
func (s *DBStore) Sync(ctx context.Context) error {
	ticker := time.NewTicker(keySyncInterval)
	defer ticker.Stop()

	for {
		if err := s.sync(ctx); err != nil {
			_ = telemetry.Err(ctx, fmt.Errorf("background sync failed: %w", err))
		} else {
			s.mu.RLock()
			hasSigning := s.signingKey.Signer != nil && s.signingKey.RotatesAt.After(time.Now())
			s.mu.RUnlock()

			if !hasSigning {
				_ = telemetry.Err(ctx, errors.New("lost signing key, triggering emergency rotation"))
				s.rotMu.Lock()

				if rotateErr := s.rotate(ctx); rotateErr != nil {
					_ = telemetry.Err(ctx, fmt.Errorf("emergency rotation failed: %w", rotateErr))
				}
				s.rotMu.Unlock()
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (s *DBStore) Rotate(ctx context.Context) error {
	for {
		s.mu.RLock()
		wakeAt := s.signingKey.RotatesAt.Add(-keyRotationLookahead)
		s.mu.RUnlock()

		timer := time.NewTimer(max(0, time.Until(wakeAt)))

		select {
		case <-timer.C:
			s.rotMu.Lock()
			err := s.rotate(ctx)
			s.rotMu.Unlock()

			if err != nil {
				_ = telemetry.Err(ctx, fmt.Errorf("background rotation failed: %w", err))

				select {
				case <-time.After(keyRotationBackoff):
				case <-ctx.Done():
					return nil
				}
			}

		case <-ctx.Done():
			timer.Stop()
			return nil
		}
	}
}

func (s *DBStore) SigningKey(_ context.Context) (Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.signingKey.Signer == nil {
		return Key{}, errors.New("jws: no active signing key")
	}

	return s.signingKey, nil
}

func (s *DBStore) PublicKeys(_ context.Context) ([]Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]Key, len(s.activeKeys))
	copy(keys, s.activeKeys)

	return keys, nil
}
