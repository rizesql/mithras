package auth

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/internal/token"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/telemetry"
)

type Logout struct {
	db  *db.Database
	clk clock.Clock
}

func NewLogout(db *db.Database, clk clock.Clock) *Logout {
	return &Logout{
		db:  db,
		clk: clk,
	}
}

func (l *Logout) Logout(ctx context.Context, rawToken string) (err error) {
	ctx, span := telemetry.Start(ctx, "auth.Logout")
	defer telemetry.End(span, &err)

	tok := token.Refresh(rawToken)
	tokHash := tok.Hash()

	sess, err := db.Query.GetSessionByTokenHash(ctx, l.db, tokHash)
	if err != nil {
		if db.IsNotFound(err) {
			return errInvalidLogoutToken("session not found")
		}

		return errLogoutSessionLookupFailed(err)
	}

	now := l.clk.Now()

	if sess.ExpiresAt.Before(now) {
		return errInvalidLogoutToken("session expired")
	}

	if sess.RevokedAt != nil {
		telemetry.Event(ctx, "auth.refresh_token_replay",
			attribute.String("session.id", sess.ID.String()),
			attribute.String("user.id", sess.UserID.String()),
		)

		revErr := db.Query.RevokeUserSessions(ctx, l.db, db.RevokeUserSessionsParams{
			Now:    &now,
			UserPk: sess.UserPk,
		})
		if revErr != nil {
			telemetry.Event(ctx, "auth.replay_revocation_failed",
				attribute.String("error", revErr.Error()),
			)
			telemetry.Err(ctx, revErr)
		}

		return errInvalidLogoutToken("token was previously revoked; anomaly detected")
	}

	revErr := db.Query.RevokeUserSessions(ctx, l.db, db.RevokeUserSessionsParams{
		Now:    &now,
		UserPk: sess.UserPk,
	})
	if revErr != nil {
		telemetry.Event(ctx, "auth.logout_revocation_failed",
			attribute.String("error", revErr.Error()),
			attribute.String("session.id", sess.ID.String()),
			attribute.String("user.id", sess.UserID.String()),
		)
		telemetry.Err(ctx, revErr)

		return errLogoutSessionRevocationFailed(revErr)
	}

	telemetry.Event(ctx, "auth.logout_success",
		attribute.String("session.id", sess.ID.String()),
		attribute.String("user.id", sess.UserID.String()),
	)

	return nil
}
