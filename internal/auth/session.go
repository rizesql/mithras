package auth

import (
	"time"

	"github.com/rizesql/mithras/internal/token"
	"github.com/rizesql/mithras/pkg/idkit"
)

type Session struct {
	ID        idkit.SessionID
	Token     token.Refresh
	ExpiresAt time.Time
}

func newSession(now time.Time, duration time.Duration) (Session, error) {
	tok, err := token.GenerateRefresh()
	if err != nil {
		return Session{}, err
	}

	return Session{
		ID:        idkit.NewSessionID(),
		Token:     tok,
		ExpiresAt: now.Add(duration),
	}, nil
}
