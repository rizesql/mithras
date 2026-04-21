package httpkit

import (
	"time"

	"github.com/spf13/pflag"
)

// Config holds the configuration for the HTTP server.
type Config struct {
	MaxRequestBodySize int64         `mapstructure:"max_request_body_size"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout"`
	RequestTimeout     time.Duration `mapstructure:"request_timeout"`
}

// DefaultConfig returns the default configuration for the HTTP server.
func DefaultConfig() Config {
	return Config{
		MaxRequestBodySize: 10 * 1024 * 1024, // 10MB
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       20 * time.Second,
		RequestTimeout:     30 * time.Second,
	}
}

// Flags returns the flag set for the HTTP server configuration.
func Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("server", pflag.ExitOnError)

	cfg := DefaultConfig()
	fs.Int64("server.max_request_body_size", cfg.MaxRequestBodySize,
		"Maximum request body size in bytes")
	fs.Duration("server.read_timeout", cfg.ReadTimeout, "Read timeout for HTTP requests")
	fs.Duration("server.write_timeout", cfg.WriteTimeout, "Write timeout for HTTP requests")
	fs.Duration("server.request_timeout", cfg.RequestTimeout, "Request timeout for HTTP requests")

	return fs
}
