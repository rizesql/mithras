package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver with database/sql
	"github.com/pressly/goose/v3"

	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

// client holds a configured database client for goose operations
type client struct {
	DB       *sql.DB
	Provider *goose.Provider
}

// close closes the underlying connection
func (c *client) close() error {
	if c != nil && c.DB != nil {
		return c.DB.Close()
	}

	return nil
}

func newClient(ctx context.Context, cfg *db.Config) (*client, error) {
	conn, err := sql.Open("pgx", cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to open migration db: %w", err)
	}

	if _, err = conn.ExecContext(ctx, "SET search_path TO "+cfg.SchemaName); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to set search path: %w", err)
	}

	migrationFS, err := fs.Sub(db.Migrations, "migrations")
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to scope migration filesystem: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		conn,
		migrationFS,
		goose.WithTableName(cfg.MigrationsTable),
		goose.WithSlog(logger.Logger),
	)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to create goose provider: %w", err)
	}

	return &client{
		DB:       conn,
		Provider: provider,
	}, nil
}
