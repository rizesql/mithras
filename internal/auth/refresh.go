package auth

import (
	"context"

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

	tok := token.Refresh(rawToken)
	tokHash := tok.Hash()

	sess, err := db.Query.GetSessionByTokenHash(ctx, r.db, tokHash)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errInvalidRefreshToken("session not found")
		}
		return nil, err
	}

	now := r.clk.Now()

	if sess.ExpiresAt.Before(now) {
		return nil, errInvalidRefreshToken("session expired")
	}

	if sess.RevokedAt != nil {
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
				attribute.String("error", revErr.Error()),
			)
			telemetry.Err(ctx, revErr)
		}

		return nil, errInvalidRefreshToken("token was previously revoked; anomaly detected")
	}
	if sess.UserStatus == db.UserStatusSuspended {
		return nil, errAccountSuspended
	}

	if sess.UserStatus == db.UserStatusLocked && sess.UserLockedUntil != nil && sess.UserLockedUntil.After(now) {
		return nil, errAccountLocked(sess.UserLockedUntil.String())
	}

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

	err = db.Tx(ctx, r.db, func(tx db.DBTX) error {
		rowsAffected, err := db.Query.RevokeSession(ctx, tx, sess.Pk)
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return errInvalidRefreshToken("concurrent revocation detected")
		}

		return db.Query.InsertSession(ctx, tx, db.InsertSessionParams{
			ID:        newSess.ID,
			UserPk:    sess.UserPk,
			TokenHash: newSess.Token.Hash(),
			UserAgent: sess.UserAgent,
			IpAddr:    sess.IpAddr,
			ExpiresAt: newSess.ExpiresAt,
		})
	})

	if err != nil {
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
