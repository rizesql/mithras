// Package routes registers the HTTP routes for the Mithras service.
package routes

import (
	"time"

	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
	"github.com/rizesql/mithras/services/mithras/platform"
	"github.com/rizesql/mithras/services/mithras/routes/docs"
	"github.com/rizesql/mithras/services/mithras/routes/openapi"
)

// Register registers the HTTP routes for the Mithras service.
func Register(srv *httpkit.Server, p *platform.Platform) {
	withLogging := middleware.WithTracing(p.Clock)
	withPanicRecovery := middleware.WithPanicRecovery()
	withTimeout := middleware.WithTimeout(time.Minute)

	mw := []httpkit.Middleware{withPanicRecovery, withLogging, withTimeout}
	srv.RegisterRoute(openapi.New(), mw...)
	srv.RegisterRoute(docs.New(), mw...)
}
