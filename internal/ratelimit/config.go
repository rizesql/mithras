package ratelimit

import (
	"time"

	"github.com/spf13/pflag"
)

type Type string

const (
	TypeRedis  Type = "redis"
	TypeMemory Type = "memory"
	TypeNoop   Type = "noop"
)

// Config represents the unified rate limiter configuration.
type Config struct {
	Type   Type         `mapstructure:"type"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Memory MemoryConfig `mapstructure:"memory"`
}

type RedisConfig struct {
	// Format: redis://[[username][:password]@][host][:port][/database]
	URL string `mapstructure:"url"`
}

type MemoryConfig struct {
	// CleanupInterval determines how often to clear expired keys in MemoryStore
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

// DefaultConfig returns the secure default configuration.
func DefaultConfig() Config {
	return Config{
		Type: TypeNoop, // Defaults to no-op for tests unless explicitly set
		Redis: RedisConfig{
			URL: "redis://localhost:6379/0",
		},
		Memory: MemoryConfig{
			CleanupInterval: time.Minute,
		},
	}
}

// Flags registers the configuration flags for the rate limiter.
func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("ratelimit", pflag.ExitOnError)
	cfg := DefaultConfig()

	f.String("ratelimit.type", string(cfg.Type),
		"Backend type for rate limiting (redis, memory, noop)",
	)
	f.String("ratelimit.redis.url", cfg.Redis.URL,
		"Redis server URL for rate limiting",
	)
	f.Duration("ratelimit.memory.cleanup_interval", cfg.Memory.CleanupInterval,
		"Interval to clean up expired in-memory rate limits",
	)

	return f
}
