package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx"
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
	if c == nil || c.DB == nil {
		return nil
	}
	return c.DB.Close()
}

func newClient(ctx context.Context, cfg *db.Config) (res *client, err error) {
	conn, err := sql.Open("pgx", cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to open migration db: %w", err)
	}

	success := false
	defer func() {
		if !success {
			err = conn.Close()
		}
	}()

	if err = conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	schema := pgx.Identifier{cfg.SchemaName}.Sanitize()
	if _, err = conn.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
		return nil, fmt.Errorf("failed to set search path: %w", err)
	}

	migrationFS, err := fs.Sub(db.Migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to access migration sub-directory: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		conn,
		migrationFS,
		goose.WithTableName(cfg.MigrationsTable),
		goose.WithSlog(logger.Logger),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create goose provider: %w", err)
	}

	success = true
	return &client{
		DB:       conn,
		Provider: provider,
	}, nil
}
