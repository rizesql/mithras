// Package routes registers the HTTP routes for the Mithras service.
package routes

import (
	"time"

	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
	"github.com/rizesql/mithras/services/mithras/platform"
	"github.com/rizesql/mithras/services/mithras/routes/docs"
	"github.com/rizesql/mithras/services/mithras/routes/openapi"
	"github.com/rizesql/mithras/services/mithras/routes/register"
)

// Register registers the HTTP routes for the Mithras service.
func Register(srv *httpkit.Server, p *platform.Platform) {
	withTelemetry := middleware.WithTelemetry()
	withPanicRecovery := middleware.WithPanicRecovery()
	withTimeout := middleware.WithTimeout(time.Minute)
	withValidation := middleware.WithValidation(p.Validator)

	withRateLimit := middleware.WithRateLimit(ratelimit.NewPolicy("global-per-ip",
		1000, time.Minute,
		ratelimit.KeyIP(),
		ratelimit.WithStore(p.RateLimit),
		ratelimit.WithBurst(),
		ratelimit.WithFailOpen(),
	))

	mw := make([]httpkit.Middleware, 0, 5)
	mw = append(mw, withPanicRecovery, withTelemetry, withTimeout, withRateLimit, withValidation)
	srv.RegisterRoute(register.New(p), append(mw, register.RateLimit(p))...)

	mw = []httpkit.Middleware{withTelemetry, withPanicRecovery}
	srv.RegisterRoute(openapi.New(), mw...)
	srv.RegisterRoute(docs.New(), mw...)

}
