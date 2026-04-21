package login

import "github.com/rizesql/mithras/internal/errkit"

var (
	errInvalidCredentials = errkit.New("invalid email or password",
		errkit.User.Auth.Code("invalid_credentials"),
		errkit.Internal("authentication failed"),
		errkit.Public("Invalid email or password."),
	)

	errMissingAuthState = errkit.New("missing authorization state",
		errkit.User.Request.Code("missing_auth_state"),
		errkit.Internal("no Auth-State cookie found; user likely hit /login directly"),
		errkit.Public("Your session has expired or the authorization request was invalid. Please try logging in from your application again."),
	)
)
