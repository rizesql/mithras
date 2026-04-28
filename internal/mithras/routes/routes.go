// Package routes registers the HTTP routes for the Mithras service.
package routes

import (
	"time"

	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/internal/mithras/routes/authorize"
	"github.com/rizesql/mithras/internal/mithras/routes/docs"
	forgotpassword "github.com/rizesql/mithras/internal/mithras/routes/forgot-password"
	"github.com/rizesql/mithras/internal/mithras/routes/jwks"
	"github.com/rizesql/mithras/internal/mithras/routes/login"
	"github.com/rizesql/mithras/internal/mithras/routes/logout"
	"github.com/rizesql/mithras/internal/mithras/routes/oas"
	"github.com/rizesql/mithras/internal/mithras/routes/openapi"
	"github.com/rizesql/mithras/internal/mithras/routes/register"
	resetpassword "github.com/rizesql/mithras/internal/mithras/routes/reset-password"
	"github.com/rizesql/mithras/internal/mithras/routes/static"
	"github.com/rizesql/mithras/internal/mithras/routes/token"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
)

// Register registers the HTTP routes for the Mithras service.
//
//nolint:funlen
func Register(srv *httpkit.Server, plt *platform.Platform) {
	withPanicRecovery := middleware.WithPanicRecovery()
	withTimeout := middleware.WithTimeout(time.Minute)
	withValidation := middleware.WithValidation(plt.Validator)

	srv.RegisterRoute(oas.New(),
		withPanicRecovery,
		withTimeout,
	)
	srv.RegisterRoute(jwks.New(plt),
		withPanicRecovery,
		withTimeout,
	)

	srv.RegisterRoute(authorize.New(plt),
		withPanicRecovery,
		withTimeout,
	)
	srv.RegisterRoute(token.New(plt),
		withPanicRecovery,
		withTimeout,
		withValidation,
	)

	srv.RegisterRoute(register.New(plt),
		withPanicRecovery,
		withTimeout,
		withValidation,
	)
	srv.RegisterRoute(login.New(plt),
		withPanicRecovery,
		withTimeout,
		withValidation,
	)
	srv.RegisterRoute(logout.New(plt),
		withPanicRecovery,
		withTimeout,
		withValidation,
	)
	srv.RegisterRoute(forgotpassword.New(plt),
		withPanicRecovery,
		withTimeout,
		withValidation,
	)
	srv.RegisterRoute(resetpassword.New(plt),
		withPanicRecovery,
		withTimeout,
		withValidation,
	)

	srv.RegisterRoute(static.New(), withPanicRecovery, withTimeout)
	srv.RegisterRoute(openapi.New(), withPanicRecovery, withTimeout)
	srv.RegisterRoute(docs.New(), withPanicRecovery, withTimeout)
}
