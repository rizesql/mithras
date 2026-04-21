package auth

import (
	"context"

	"github.com/rizesql/mithras/internal/token"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

const (
	AuthSessionCookie = "mithras_session"
)

type Verifier struct {
	db  *db.Database
	clk clock.Clock
}

func NewVerifier(d *db.Database, clk clock.Clock) *Verifier {
	return &Verifier{db: d, clk: clk}
}

func (v *Verifier) Verify(ctx context.Context, rawToken string) (db.User, error) {
	ctx, span := telemetry.Start(ctx, "auth.Verify")
	defer telemetry.End(span, nil)

	tok := token.Refresh(rawToken)
	hash := tok.Hash()

	sess, err := db.Query.GetSessionByTokenHash(ctx, v.db, hash)
	if err != nil {
		if db.IsNotFound(err) {
			return db.User{}, errSessionNotFound
		}
		return db.User{}, err
	}

	now := v.clk.Now()
	if sess.ExpiresAt.Before(now) {
		return db.User{}, errSessionExpired
	}
	if sess.RevokedAt != nil {
		return db.User{}, errSessionRevoked
	}

	if sess.UserStatus == db.UserStatusSuspended {
		return db.User{}, errAccountSuspended
	}

	if sess.UserStatus == db.UserStatusLocked && sess.UserLockedUntil != nil && sess.UserLockedUntil.After(now) {
		return db.User{}, errAccountLocked(sess.UserLockedUntil.String())
	}

	usr, err := db.Query.GetUserByPk(ctx, v.db, sess.UserPk)
	if err != nil {
		return db.User{}, err
	}

	return usr, nil
}

func User(c *httpkit.Context) db.User {
	if u, ok := c.User().(db.User); ok {
		return u
	}
	return db.User{}
}
