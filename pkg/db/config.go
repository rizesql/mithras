package db

import "time"

type Config struct {
	URI                         string        `mapstructure:"uri"`
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
		MaxConnections:        20,
		MinConnections:        0,
		MinIdleConnections:    1,
		MaxConnectionLifetime: 300 * time.Second,
		MaxConnectionIdleTime: 60 * time.Second,
	}
}
