package config

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

// Config holds the configuration for the Mithras service.
type Config struct {
	Issuer    string            `mapstructure:"issuer"`
	HTTPPort  string            `mapstructure:"http_port"`
	Server    httpkit.Config    `mapstructure:"server"`
	Logs      telemetry.Logs    `mapstructure:"logs"`
	Tracing   telemetry.Tracing `mapstructure:"tracing"`
	Metrics   telemetry.Metrics `mapstructure:"metrics"`
	DB        db.Config         `mapstructure:"db"`
	RateLimit ratelimit.Config  `mapstructure:"ratelimit"`
	Auth      auth.Config       `mapstructure:"auth"`
}

// DefaultConfig returns the default configuration for the Mithras service.
func DefaultConfig() Config {
	return Config{
		Issuer:    "http://localhost:8080",
		HTTPPort:  "8080",
		Server:    httpkit.DefaultConfig(),
		Logs:      telemetry.DefaultLogsConfig(),
		Tracing:   telemetry.DefaultTracingConfig(),
		Metrics:   telemetry.DefaultMetricsConfig(),
		RateLimit: ratelimit.DefaultConfig(),
		Auth:      auth.DefaultConfig(),
	}
}

// Flags returns the flag set for the Mithras service.
func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ExitOnError)

	cfg := DefaultConfig()
	f.String("issuer", cfg.Issuer, "Issuer URL to use for JWTs and well-known endpoints")
	f.String("http_port", cfg.HTTPPort, "HTTP port to listen on")
	f.AddFlagSet(httpkit.Flags())
	f.AddFlagSet(telemetry.LogsFlags())
	f.AddFlagSet(telemetry.TracingFlags())
	f.AddFlagSet(telemetry.MetricsFlags())
	f.AddFlagSet(db.Flags())
	f.AddFlagSet(ratelimit.Flags())
	f.AddFlagSet(auth.Flags())

	return f
}

func bytesBase64Hook() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}

		if to.Kind() != reflect.Slice || to.Elem().Kind() != reflect.Uint8 {
			return data, nil
		}

		s, ok := data.(string)
		if !ok || s == "" {
			return data, nil
		}

		decoded, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 bytes: %w", err)
		}

		return decoded, nil
	}
}

func Load(v *viper.Viper) (Config, error) {
	cfg := DefaultConfig()

	err := v.Unmarshal(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			bytesBase64Hook(),
			mapstructure.TextUnmarshallerHookFunc(),
		),
	))
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal mithras config: %w", err)
	}

	return cfg, nil
}
