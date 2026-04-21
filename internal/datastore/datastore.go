package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/rizesql/mithras/pkg/runtime"
	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

func Migrate(ctx context.Context, cfg *Config) error {
	logger.Info("datastore.migrate.starting")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rt := runtime.New(ctx)
	defer rt.Recover()

	client, err := newClient(ctx, &cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to create migration client: %w", err)
	}

	defer func() {
		if closeErr := client.close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close migration client: %w", closeErr))
		}
	}()

	rt.Go(func(rtCtx context.Context) error {
		defer cancel()

		if _, err := client.Provider.Up(rtCtx); err != nil {
			return fmt.Errorf("failed to run migrations: %w", sanitizeSQLError(err))
		}

		return nil
	})

	if err := rt.Run(ctx); err != nil {
		return fmt.Errorf("migrate failed: %w", err)
	}

	logger.Info("datastore.migrate.stopped")

	return nil
}

func Status(ctx context.Context, cfg *Config) (*MigrationStatus, error) {
	logger.Info("datastore.status.starting")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rt := runtime.New(ctx)
	defer rt.Recover()

	client, err := newClient(ctx, &cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration client: %w", err)
	}

	defer func() {
		if closeErr := client.close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close migration client: %w", closeErr))
		}
	}()

	var ms *MigrationStatus

	rt.Go(func(rtCtx context.Context) error {
		defer cancel()

		if ms, err = client.status(rtCtx); err != nil {
			return fmt.Errorf("failed to get migration status: %w", err)
		}

		return nil
	})

	if err := rt.Run(ctx); err != nil {
		return nil, fmt.Errorf("status failed: %w", err)
	}

	logger.Info("datastore.status.stopped")

	return ms, nil
}
