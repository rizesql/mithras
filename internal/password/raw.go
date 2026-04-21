package password

import (
	"fmt"
	"unicode"

	"github.com/rizesql/mithras/internal/errkit"
)

const (
	minLength = 8
)

var (
	ErrWeak = errkit.New("password does not meet complexity requirements",
		errkit.WithCode(errkit.User.Request.Code("invalid_password")),
		errkit.Internal("password complexity validation failed"),
		errkit.Public("Password must contain at least one uppercase, one lowercase, one digit, and one special character."),
	)
	ErrTooShort = errkit.New(fmt.Sprintf("password is too short; minimum length is %d characters", minLength),
		errkit.WithCode(errkit.User.Request.Code("invalid_password")),
		errkit.Internal("password length validation failed"),
		errkit.Publicf("Password is too short; minimum length is %d characters.", minLength),
	)
)

type Raw struct {
	value string
}

func (Raw) String() string { return "[REDACTED]" }

func New(raw string) (Raw, error) {
	if len(raw) < minLength {
		return Raw{}, ErrTooShort
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range raw {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return Raw{}, ErrWeak
	}

	return Raw{value: raw}, nil
}
