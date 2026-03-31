// Package mithras provides the Mithras service.
package mithras

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/logger"
	"github.com/rizesql/mithras/pkg/runtime"
	"github.com/rizesql/mithras/pkg/telemetry"
	"github.com/rizesql/mithras/services/datastore"
	"github.com/rizesql/mithras/services/mithras/platform"
	"github.com/rizesql/mithras/services/mithras/routes"
)

// Run bootstraps the Mithras service.
func Run(ctx context.Context, cfg *Config) error {
	logger.Configure(cfg.Logs)
	logger.Info("mithras.started")

	rt := runtime.New(ctx)
	defer rt.Recover()

	shutdown, err := telemetry.ConfigureTracing(ctx, &cfg.Tracing)
	if err != nil {
		return fmt.Errorf("unable to configure tracing: %w", err)
	}
	rt.DeferFunc(shutdown)

	shutdown, err = telemetry.ConfigureMetrics(ctx, &cfg.Metrics)
	if err != nil {
		return fmt.Errorf("unable to configure metrics: %w", err)
	}
	rt.DeferFunc(shutdown)

	clk := clock.System

	v, err := validator.New()
	if err != nil {
		return fmt.Errorf("unable to create validator: %w", err)
	}

	if err := datastore.CheckPendingMigrations(ctx, &cfg.DB); err != nil {
		return fmt.Errorf("migrations pending (run 'mithras datastore migrate'): %w", err)
	}
	database, err := db.New(ctx, &cfg.DB)
	if err != nil {
		return fmt.Errorf("unable to create database: %w", err)
	}
	rt.Defer(database.Close)

	rl, err := ratelimit.New(ctx, &cfg.RateLimit)
	if err != nil {
		return fmt.Errorf("unable to create rate limit store: %w", err)
	}

	plt := platform.New(v, clk, database, rl)

	srv := httpkit.New(httpkit.Dependencies{Clock: plt.Clock}, cfg.Server)
	rt.RegisterHealth(srv.Mux())
	rt.DeferFunc(srv.Shutdown)

	routes.Register(srv, plt)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTPPort))
	if err != nil {
		logger.Error("tcp.listen_failed", "error", err, "port", cfg.HTTPPort)
		return fmt.Errorf("unable to listen on port %d: %w", cfg.HTTPPort, err)
	}

	rt.Go(func(ctx context.Context) error {
		err := srv.Serve(ctx, ln)
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server.serve_failed", "error", err)
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	})

	if err := rt.Run(ctx, runtime.WithTimeout(time.Minute)); err != nil {
		logger.Error("runtime.shutdown_failed", "error", err)
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("mithras.stopped")
	return nil
}
