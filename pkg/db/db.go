package db

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizesql/mithras/pkg/retry"
	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

//go:embed migrations/*.sql
var Migrations embed.FS

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Queries struct{}

var Query Querier = &Queries{}

type Database struct {
	*pgxpool.Pool
	maxRetries int
}

func New(ctx context.Context, cfg *Config) (*Database, error) {
	logger.Info("db.starting")

	poolCfg, err := applyConfig(cfg)
	if err != nil {
		return nil, err
	}

	pool, err := initPool(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	logger.Info("db.pool.ping")

	err = ping(ctx, pool)
	if err != nil {
		pool.Close()
		return nil, err
	}

	logger.Info("db.pool.connected")

	return &Database{
		Pool:       pool,
		maxRetries: cfg.MaxRetries,
	}, nil
}

func applyConfig(cfg *Config) (*pgxpool.Config, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URI: %w", err)
	}

	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTrimSQLInSpanName())

	if cfg.SchemaName != "" {
		if poolCfg.ConnConfig.RuntimeParams == nil {
			poolCfg.ConnConfig.RuntimeParams = make(map[string]string)
		}

		poolCfg.ConnConfig.RuntimeParams["search_path"] = cfg.SchemaName
	}

	if cfg.MaxConnections > 0 {
		poolCfg.MaxConns = cfg.MaxConnections
	}

	if cfg.MinConnections > 0 {
		poolCfg.MinConns = cfg.MinConnections
	}

	if cfg.MinIdleConnections > 0 {
		poolCfg.MinIdleConns = cfg.MinIdleConnections
	}

	poolCfg.MaxConnIdleTime = cfg.MaxConnectionIdleTime
	poolCfg.MaxConnLifetime = cfg.MaxConnectionLifetime

	if cfg.MaxConnectionLifetimeJitter > 0 {
		poolCfg.MaxConnLifetimeJitter = cfg.MaxConnectionLifetimeJitter
	} else if cfg.MaxConnectionLifetime > 0 {
		poolCfg.MaxConnLifetimeJitter = time.Duration(0.2 * float64(cfg.MaxConnectionLifetime))
	}

	if cfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	if cfg.ConnectTimeout > 0 {
		poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	}

	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	return poolCfg, nil
}

func initPool(ctx context.Context, poolCfg *pgxpool.Config) (*pgxpool.Pool, error) {
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(initCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

func ping(ctx context.Context, pool *pgxpool.Pool) error {
	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	policy := retry.New(retry.Backoff(retry.DefaultExpBackoff()))

	err := retry.Do(pingCtx,
		policy,
		func(ctx context.Context) error {
			attemptCtx, attemptCancel := context.WithTimeout(ctx, 2*time.Second)
			defer attemptCancel()

			if err := pool.Ping(attemptCtx); err != nil {
				return fmt.Errorf("database ping failed: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to verify database connection: %w", err)
	}

	return nil
}

func (db *Database) Close() error {
	if db != nil && db.Pool != nil {
		db.Pool.Close()
	}

	return nil
}

func (db *Database) IsReady(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return false, err
	}

	return true, nil
}
