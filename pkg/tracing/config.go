package tracing

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
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

// Config holds all configuration options for the tracing package.
type Config struct {
	// Level is the minimum log level to emit. Defaults to slog.LevelInfo.
	// Accepts "debug", "info", "warn"/"warning", "error" from the environment.
	Level slog.Level `mapstructure:"level"`

	// Enabled controls whether the tracer emits any output. Defaults to true.
	Enabled bool `mapstructure:"enabled"`

	// Format selects the output format. Defaults to FormatJSON.
	// Accepts "json" or "text" from the environment.
	Format Format `mapstructure:"format"`

	// Sampler cannot be expressed in the environment and must be set
	// programmatically after decoding. Defaults to [AlwaysSample].
	Sampler Sampler `mapstructure:"-"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Level:   slog.LevelInfo,
		Enabled: true,
		Format:  FormatJSON,
		Sampler: AlwaysSample{},
	}
}

// Flags returns a pflag.FlagSet for configuring tracing settings.
func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("tracing", pflag.ExitOnError)

	cfg := DefaultConfig()
	f.String("tracing.level", cfg.Level.String(), "Tracing level")
	f.Bool("tracing.enabled", cfg.Enabled, "Enable tracing")
	f.String("tracing.format", cfg.Format.String(), "Tracing format")
	return f
}

// DecodeHook returns a mapstructure.DecodeHookFunc that teaches mapstructure
// how to convert string values from Viper into [slog.Level] and [Format].
func DecodeHook() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		s := data.(string)

		switch to {
		case reflect.TypeFor[slog.Level]():
			return parseLevel(s)
		case reflect.TypeFor[Format]():
			return parseFormat(s)
		}

		return data, nil
	}
}

func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("tracing: unknown level %q (want debug|info|warn|error)", s)
	}
}

func parseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "json":
		return FormatJSON, nil
	case "text":
		return FormatText, nil
	default:
		return 0, fmt.Errorf("tracing: unknown format %q (want json|text)", s)
	}
}
