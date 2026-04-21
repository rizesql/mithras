package register

import "github.com/rizesql/mithras/internal/errkit"

var (
	errRegistrationFailed = func(err error) error {
		return errkit.Wrap(err,
			errkit.User.Request.Code("registration_failed"),
			errkit.Internal("registration failed"),
			errkit.Public(err.Error()),
		)
	}

	errMissingAuthState = errkit.New("missing authorization state",
		errkit.User.Request.Code("missing_auth_state"),
		errkit.Internal("no Auth-State cookie found; user likely hit /register directly"),
		errkit.Public("Your session has expired or the authorization request was invalid. Please try logging in from your application again."),
	)
)
