package logger

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/pflag"
)

// Format controls the log output format.
type Format int

const (
	// FormatJSON - JSON (default)
	FormatJSON Format = iota
	// FormatText - human-readable text
	FormatText
)

func (f Format) String() string {
	switch f {
	case FormatJSON:
		return "json"
	case FormatText:
		return "text"
	default:
		return "unknown"
	}
}

func (f *Format) UnmarshalText(text []byte) error {
	switch strings.ToLower(strings.TrimSpace(string(text))) {
	case "", "json":
		*f = FormatJSON
	case "text":
		*f = FormatText
	default:
		return fmt.Errorf("logger: unknown format %q (want json|text)", text)
	}
	return nil
}

// Config holds all configuration options for the logger package.
type Config struct {
	// Enabled controls whether the tracer emits any output. Defaults to true.
	Enabled bool `mapstructure:"enabled"`

	// Level is the minimum log level to emit. Defaults to slog.LevelInfo.
	// Accepts "debug", "info", "warn"/"warning", "error" from the environment.
	Level slog.Level `mapstructure:"level"`

	// Format selects the output format. Defaults to FormatJSON.
	// Accepts "json" or "text" from the environment.
	Format Format `mapstructure:"format"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Level:   slog.LevelInfo,
		Format:  FormatJSON,
	}
}

// Flags returns a pflag.FlagSet for configuring logger settings.
func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("logs", pflag.ExitOnError)

	cfg := DefaultConfig()
	f.String("logs.level", cfg.Level.String(), "logs level")
	f.Bool("logs.enabled", cfg.Enabled, "Enable logs")
	f.String("logs.format", cfg.Format.String(), "logs format")
	return f
}
