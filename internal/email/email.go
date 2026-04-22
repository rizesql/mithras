package email

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"

	"github.com/rizesql/mithras/internal/errkit"
)

var (
	ErrInvalidEmail = errkit.New("invalid email address",
		errkit.WithCode(errkit.User.Request.Code("invalid_email")),
		errkit.Internal("email validation failed"),
		errkit.Public("Invalid email address format."),
	)

	// We use \x60 to safely represent the backtick character inside a raw string.
	//nolint:lll
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_\x60{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`)
)

// Address represents a parsed, PII-safe email address.
type Address struct {
	raw    string
	local  string
	domain string
}

// Parse takes a raw email string, validates its basic structure,
// and splits it into the Local Part and Domain.
func Parse(raw string) (Address, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))

	// Validate against the exact DB Schema Regex
	if !emailRegex.MatchString(raw) {
		return Address{}, ErrInvalidEmail
	}

	idx := strings.LastIndex(raw, "@")
	if idx <= 0 || idx == len(raw)-1 {
		return Address{}, ErrInvalidEmail
	}

	return Address{
		raw:    raw,
		local:  raw[:idx],
		domain: raw[idx+1:],
	}, nil
}

// Raw returns the fully unmasked, original email address.
// Use this ONLY when passing to external APIs.
func (e Address) Raw() string { return e.raw }

// Local returns the local part of the email.
func (e Address) Local() string { return e.local }

// Domain returns the domain part of the email.
func (e Address) Domain() string { return e.domain }

// String implements the fmt.Stringer interface.
// It automatically masks the local part to prevent PII leaks in logs,
// traces, and console output.
func (e Address) String() string {
	if e.raw == "" {
		return ""
	}

	length := len(e.local)

	var maskedLocal string

	switch {
	case length <= 2:
		maskedLocal = "***"
	default:
		maskedLocal = string(e.local[0]) + "***" + string(e.local[length-1])
	}

	return maskedLocal + "@" + e.domain
}

// UnmarshalText implements encoding.TextUnmarshaler (For incoming JSON).
func (e *Address) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}

	*e = parsed

	return nil
}

// MarshalText implements encoding.TextMarshaler (For outgoing JSON).
func (e Address) MarshalText() ([]byte, error) {
	return []byte(e.raw), nil
}

func (e Address) Value() (driver.Value, error) {
	if e.raw == "" {
		return nil, ErrInvalidEmail
	}

	return e.raw, nil
}

func (e *Address) Scan(src any) error {
	if src == nil {
		return ErrInvalidEmail
	}

	var source string

	switch v := src.(type) {
	case string:
		source = v
	case []byte:
		source = string(v)
	default:
		return fmt.Errorf("incompatible type for email Address: %T", src)
	}

	parsed, err := Parse(source)
	if err != nil {
		return err
	}

	*e = parsed

	return nil
}
