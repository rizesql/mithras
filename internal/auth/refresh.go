package auth

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/internal/token"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/telemetry"
)

type Refresh struct {
	db  *db.Database
	clk clock.Clock
	iss *token.Issuer
	cfg *Config
}

func NewRefresh(d *db.Database, clk clock.Clock, iss *token.Issuer, cfg *Config) *Refresh {
	return &Refresh{
		db:  d,
		clk: clk,
		iss: iss,
		cfg: cfg,
	}
}

type RefreshResponse struct {
	AccessToken  token.Access
	RefreshToken token.Refresh
}

func (r *Refresh) Refresh(ctx context.Context, rawToken string) (resp *RefreshResponse, err error) {
	ctx, span := telemetry.Start(ctx, "auth.Refresh")
	defer telemetry.End(span, &err)

	sess, err := r.validateSession(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	now := r.clk.Now()
	if err := r.handleReplayIfNecessary(ctx, sess, now); err != nil {
		return nil, err
	}

	if err := checkUserStatus(sess.UserStatus, sess.UserLockedUntil, now); err != nil {
		return nil, err
	}

	return r.performRefresh(ctx, sess, now)
}

func (r *Refresh) validateSession(
	ctx context.Context,
	rawToken string,
) (db.GetSessionByTokenHashRow, error) {
	tok := token.Refresh(rawToken)
	tokHash := tok.Hash()

	sess, err := db.Query.GetSessionByTokenHash(ctx, r.db, tokHash)
	if err != nil {
		if db.IsNotFound(err) {
			return db.GetSessionByTokenHashRow{}, errInvalidRefreshToken("session not found")
		}
		return db.GetSessionByTokenHashRow{}, err
	}

	if sess.ExpiresAt.Before(r.clk.Now()) {
		return db.GetSessionByTokenHashRow{}, errInvalidRefreshToken("session expired")
	}

	return sess, nil
}

func (r *Refresh) handleReplayIfNecessary(
	ctx context.Context,
	sess db.GetSessionByTokenHashRow,
	now time.Time,
) error {
	if sess.RevokedAt == nil {
		return nil
	}

	telemetry.Event(ctx, "auth.refresh_token_replay",
		attribute.String("session.id", sess.ID.String()),
		attribute.String("user.id", sess.UserID.String()),
	)

	revErr := db.Query.RevokeUserSessions(ctx, r.db, db.RevokeUserSessionsParams{
		Now:    &now,
		UserPk: sess.UserPk,
	})
	if revErr != nil {
		telemetry.Event(ctx, "auth.replay_revocation_failed",
			attribute.String("error", telemetry.Err(ctx, revErr).Error()),
		)
	}

	return errInvalidRefreshToken("token was previously revoked; anomaly detected")
}

func (r *Refresh) performRefresh(
	ctx context.Context,
	sess db.GetSessionByTokenHashRow,
	now time.Time,
) (*RefreshResponse, error) {
	newSess, genErr := newSession(now, r.cfg.RefreshTokenDuration)
	if genErr != nil {
		return nil, errRefreshTokenGenerationFailed(genErr)
	}

	accessToken, err := r.iss.Issue(ctx, token.IssueConfig{
		Subject:  sess.UserID.String(),
		IssuedAt: now,
		Duration: r.cfg.AccessTokenDuration,
	})
	if err != nil {
		return nil, errRefreshTokenSigningFailed(err)
	}

	if err := r.rotateSession(ctx, sess, &newSess); err != nil {
		return nil, err
	}

	telemetry.Attr(ctx,
		attribute.Bool("auth.success", true),
		attribute.String("user.id", sess.UserID.String()),
		attribute.String("session.id", newSess.ID.String()),
	)

	return &RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newSess.Token,
	}, nil
}

func (r *Refresh) rotateSession(
	ctx context.Context,
	oldSess db.GetSessionByTokenHashRow,
	newSess *Session,
) error {
	return db.Tx(ctx, r.db, func(tx db.DBTX) error {
		rowsAffected, err := db.Query.RevokeSession(ctx, tx, oldSess.Pk)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return errInvalidRefreshToken("concurrent revocation detected")
		}

		return db.Query.InsertSession(ctx, tx, db.InsertSessionParams{
			ID:        newSess.ID,
			UserPk:    oldSess.UserPk,
			TokenHash: newSess.Token.Hash(),
			UserAgent: oldSess.UserAgent,
			IpAddr:    oldSess.IpAddr,
			ExpiresAt: newSess.ExpiresAt,
		})
	})
}
