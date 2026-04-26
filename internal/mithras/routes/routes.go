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
	"github.com/rizesql/mithras/internal/ratelimit"
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

	withRateLimit := middleware.WithRateLimit(ratelimit.NewPolicy("global-per-ip",
		1000, time.Minute,
		ratelimit.KeyIP(),
		ratelimit.WithStore(plt.RateLimit),
		ratelimit.WithBurst(),
		ratelimit.WithFailOpen(),
	))

	srv.RegisterRoute(oas.New(),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
		jwks.RateLimit(plt),
	)
	srv.RegisterRoute(jwks.New(plt),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
		jwks.RateLimit(plt),
	)

	srv.RegisterRoute(authorize.New(plt),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
	)
	srv.RegisterRoute(token.New(plt),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
		withValidation,
	)

	srv.RegisterRoute(register.New(plt),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
		register.RateLimit(plt),
		withValidation,
	)
	srv.RegisterRoute(login.New(plt),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
		login.RateLimit(plt),
		withValidation,
	)
	srv.RegisterRoute(logout.New(plt),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
		logout.RateLimit(plt),
		withValidation,
	)
	srv.RegisterRoute(forgotpassword.New(plt),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
		forgotpassword.RateLimit(plt),
		withValidation,
	)
	srv.RegisterRoute(resetpassword.New(plt),
		withPanicRecovery,
		withTimeout,
		withRateLimit,
		resetpassword.RateLimit(plt),
		withValidation,
	)

	srv.RegisterRoute(static.New(), withPanicRecovery, withTimeout)
	srv.RegisterRoute(openapi.New(), withPanicRecovery, withTimeout)
	srv.RegisterRoute(docs.New(), withPanicRecovery, withTimeout)
}
