// Package mithras provides the Mithras service.
package mithras

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/internal/datastore"
	"github.com/rizesql/mithras/internal/jws"
	"github.com/rizesql/mithras/internal/mithras/config"
	"github.com/rizesql/mithras/internal/mithras/platform"
	"github.com/rizesql/mithras/internal/mithras/routes"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/internal/token"
	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/runtime"
	"github.com/rizesql/mithras/pkg/telemetry"
	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

// Run bootstraps the Mithras service.
func Run(ctx context.Context, cfg *config.Config) error {
	rt := runtime.New(ctx)
	defer rt.Recover()

	if err := configureTelemetry(ctx, rt, cfg); err != nil {
		return err
	}

	logger.Info("mithras.starting")

	plt, err := initPlatform(ctx, rt, cfg)
	if err != nil {
		return err
	}

	if err := startServer(ctx, rt, plt, cfg); err != nil {
		return err
	}

	if err := rt.Run(ctx, runtime.WithTimeout(time.Minute)); err != nil {
		logger.Error("runtime.shutdown_failed", "error", err)
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("mithras.stopped")

	return nil
}

func initPlatform(
	ctx context.Context,
	rt *runtime.Runtime,
	cfg *config.Config,
) (*platform.Platform, error) {
	clk := clock.System

	valid, err := validator.New()
	if err != nil {
		return nil, fmt.Errorf("unable to create validator: %w", err)
	}

	database, err := initDatabase(ctx, rt, cfg)
	if err != nil {
		return nil, err
	}

	rl, err := ratelimit.New(ctx, &cfg.RateLimit)
	if err != nil {
		return nil, fmt.Errorf("unable to create rate limit store: %w", err)
	}

	jwsStore, err := jws.NewDBStore(ctx, database, cfg.Auth.KEK, jws.EdDSA{})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize jws store: %w", err)
	}

	rt.Go(jwsStore.Sync)
	rt.Go(jwsStore.Rotate)

	issuer := token.NewIssuer(jwsStore, cfg.Issuer)

	oauth2, err := auth.NewOAuth2(database, clk, cfg.Auth.KEK)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize oauth2: %w", err)
	}

	pr := auth.NewPasswordReset(database, clk)

	return platform.New(valid, clk, database, rl, jwsStore, &issuer, cfg, oauth2, pr), nil
}

func initDatabase(
	ctx context.Context,
	rt *runtime.Runtime,
	cfg *config.Config,
) (*db.Database, error) {
	if err := datastore.CheckPendingMigrations(ctx, &cfg.DB); err != nil {
		return nil, fmt.Errorf("migrations pending (run 'mithras datastore migrate'): %w", err)
	}

	database, err := db.New(ctx, &cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("unable to create database: %w", err)
	}

	rt.Defer(database.Close)

	return database, nil
}

func startServer(
	ctx context.Context,
	rt *runtime.Runtime,
	plt *platform.Platform,
	cfg *config.Config,
) error {
	srv := httpkit.New(httpkit.Dependencies{Clock: plt.Clock}, cfg.Server)
	rt.RegisterHealth(srv.Mux())
	rt.DeferFunc(srv.Shutdown)

	routes.Register(srv, plt)

	ln, err := new(net.ListenConfig{}).Listen(ctx, "tcp", ":"+cfg.HTTPPort)
	if err != nil {
		logger.Error("tcp.listen_failed", "error", err, "port", cfg.HTTPPort)
		return fmt.Errorf("unable to listen on port %s: %w", cfg.HTTPPort, err)
	}

	rt.Go(func(ctx context.Context) error {
		err := srv.Serve(ctx, ln)
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server.serve_failed", "error", err)
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	})

	return nil
}

func configureTelemetry(ctx context.Context, rt *runtime.Runtime, cfg *config.Config) error {
	shutdown, err := telemetry.ConfigureLogs(ctx, &cfg.Logs)
	if err != nil {
		return fmt.Errorf("unable to configure logs: %w", err)
	}
	rt.DeferFunc(shutdown)

	shutdown, err = telemetry.ConfigureTracing(ctx, &cfg.Tracing)
	if err != nil {
		return fmt.Errorf("unable to configure tracing: %w", err)
	}
	rt.DeferFunc(shutdown)

	shutdown, err = telemetry.ConfigureMetrics(ctx, &cfg.Metrics)
	if err != nil {
		return fmt.Errorf("unable to configure metrics: %w", err)
	}
	rt.DeferFunc(shutdown)

	return nil
}
