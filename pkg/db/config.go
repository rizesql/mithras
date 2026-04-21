package db

import (
	"time"

	"github.com/spf13/pflag"
)

type Config struct {
	URI                         string        `mapstructure:"uri"`
	MigrationsTable             string        `mapstructure:"migrations_table"`
	SchemaName                  string        `mapstructure:"schema_name"`
	MaxConnections              int32         `mapstructure:"max_connections"`
	MinConnections              int32         `mapstructure:"min_connections"`
	MinIdleConnections          int32         `mapstructure:"min_idle_connections"`
	MaxConnectionLifetime       time.Duration `mapstructure:"max_connection_lifetime"`
	MaxConnectionIdleTime       time.Duration `mapstructure:"max_connection_idle_time"`
	HealthCheckPeriod           time.Duration `mapstructure:"health_check_period"`
	MaxConnectionLifetimeJitter time.Duration `mapstructure:"max_connection_lifetime_jitter"`
	ConnectTimeout              time.Duration `mapstructure:"connect_timeout"`
	MaxRetries                  int           `mapstructure:"max_retries"`
}

func DefaultConfig() Config {
	return Config{
		MigrationsTable:       "mithras_schema_migrations",
		SchemaName:            "public",
		MaxConnections:        20,
		MinConnections:        0,
		MinIdleConnections:    1,
		MaxConnectionLifetime: 300 * time.Second,
		MaxConnectionIdleTime: 60 * time.Second,
		MaxRetries:            3,
	}
}

func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("db", pflag.ExitOnError)

	cfg := DefaultConfig()
	fs.String("db.uri", cfg.URI, "Database URI")
	fs.String("db.migrations_table", cfg.MigrationsTable, "Custom goose migrations table name")
	fs.String("db.schema_name", cfg.SchemaName, "PostgreSQL search_path isolation schema")
	fs.Int32("db.max_connections", cfg.MaxConnections, "Maximum number of connections")
	fs.Int32("db.min_connections", cfg.MinConnections, "Minimum number of connections")
	fs.Int32("db.min_idle_connections", cfg.MinIdleConnections, "Minimum number of idle connections")
	fs.Duration("db.max_connection_lifetime", cfg.MaxConnectionLifetime, "Maximum connection lifetime")
	fs.Duration("db.max_connection_idle_time", cfg.MaxConnectionIdleTime,
		"Maximum connection idle time")
	fs.Duration("db.health_check_period", cfg.HealthCheckPeriod, "Health check period")
	fs.Duration("db.max_connection_lifetime_jitter", cfg.MaxConnectionLifetimeJitter,
		"Max connection lifetime jitter")
	fs.Duration("db.connect_timeout", cfg.ConnectTimeout, "Connect timeout")
	fs.Int("db.max_retries", cfg.MaxRetries, "Max retries")

	return fs
}
