package auth

import (
	"github.com/rizesql/mithras/internal/email"
	"github.com/rizesql/mithras/internal/password"
)

func parseCredentials(rawEmail, rawPassword string) (email.Address, password.Raw, error) {
	addr, err := email.Parse(rawEmail)
	if err != nil {
		return email.Address{}, password.Raw{}, err
	}

	pwd, err := password.New(rawPassword)
	if err != nil {
		return email.Address{}, password.Raw{}, err
	}

	return addr, pwd, nil
}
