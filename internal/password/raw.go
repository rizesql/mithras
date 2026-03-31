package password

import (
	"errors"
	"fmt"
	"unicode"
)

const (
	minLength = 8
)

var ErrWeak = errors.New("password does not meet complexity requirements")
var ErrTooShort = fmt.Errorf("password is too short; minimum length is %d characters", minLength)

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
		return Raw{}, errors.New("password must contain at least one uppercase, one lowercase, one digit, and one special character")
	}

	return Raw{value: raw}, nil
}
