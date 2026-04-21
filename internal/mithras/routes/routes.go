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

	htmlMw := []httpkit.Middleware{withPanicRecovery, withTimeout, withRateLimit}

	apiMw := make([]httpkit.Middleware, 0, len(htmlMw)+1)
	apiMw = append(apiMw, htmlMw...)
	apiMw = append(apiMw, withValidation)

	srv.RegisterRoute(oas.New(), append(apiMw, jwks.RateLimit(plt))...)
	srv.RegisterRoute(jwks.New(plt), append(apiMw, jwks.RateLimit(plt))...)

	srv.RegisterRoute(authorize.New(plt), htmlMw...)
	srv.RegisterRoute(token.New(plt), apiMw...)

	srv.RegisterRoute(login.New(plt), append(apiMw, login.RateLimit(plt))...)
	srv.RegisterRoute(register.New(plt), append(apiMw, register.RateLimit(plt))...)
	srv.RegisterRoute(forgotpassword.New(plt), append(apiMw, forgotpassword.RateLimit(plt))...)
	srv.RegisterRoute(resetpassword.New(plt), append(apiMw, resetpassword.RateLimit(plt))...)

	srv.RegisterRoute(static.New(), withPanicRecovery, withTimeout)
	srv.RegisterRoute(openapi.New(), withPanicRecovery, withTimeout)
	srv.RegisterRoute(docs.New(), withPanicRecovery, withTimeout)
}
