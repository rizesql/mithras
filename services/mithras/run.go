// Package mithras provides the Mithras service.
package mithras

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/runtime"
	"github.com/rizesql/mithras/pkg/tracing"
	"github.com/rizesql/mithras/services/mithras/platform"
	"github.com/rizesql/mithras/services/mithras/routes"
)

// Run bootstraps the Mithras service.
func Run(ctx context.Context, cfg *Config) error {
	rt := runtime.New(ctx)
	defer rt.Recover()

	clk := clock.System

	v, err := validator.New()
	if err != nil {
		return fmt.Errorf("unable to create validator: %w", err)
	}

	plt := platform.New(v, clk)

	srv := httpkit.New(httpkit.Dependencies{Clock: plt.Clock}, cfg.Server)
	rt.RegisterHealth(srv.Mux())
	rt.DeferFunc(srv.Shutdown)

	routes.Register(srv, plt)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTPPort))
	if err != nil {
		tracing.Error("tcp.listen_failed", "error", err, "port", cfg.HTTPPort)
		return fmt.Errorf("unable to listen on port %d: %w", cfg.HTTPPort, err)
	}

	rt.Go(func(ctx context.Context) error {
		err := srv.Serve(ctx, ln)
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
			tracing.Error("server.serve_failed", "error", err)
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	})

	if err := rt.Run(runtime.WithTimeout(time.Minute)); err != nil {
		tracing.Error("runtime.shutdown_failed", "error", err)
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	tracing.Info("mithras.server_stopped")
	return nil
}
