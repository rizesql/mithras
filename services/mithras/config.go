package mithras

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/logger"
	"github.com/rizesql/mithras/pkg/telemetry"
)

// Config holds the configuration for the Mithras service.
type Config struct {
	HTTPPort  int               `mapstructure:"http_port"`
	Server    httpkit.Config    `mapstructure:"server"`
	Logs      logger.Config     `mapstructure:"logs"`
	Tracing   telemetry.Tracing `mapstructure:"tracing"`
	Metrics   telemetry.Metrics `mapstructure:"metrics"`
	DB        db.Config         `mapstructure:"db"`
	RateLimit ratelimit.Config  `mapstructure:"ratelimit"`
}

// DefaultConfig returns the default configuration for the Mithras service.
func DefaultConfig() Config {
	return Config{
		HTTPPort:  8080,
		Server:    httpkit.DefaultConfig(),
		Logs:      logger.DefaultConfig(),
		Tracing:   telemetry.DefaultTracingConfig(),
		Metrics:   telemetry.DefaultMetricsConfig(),
		RateLimit: ratelimit.DefaultConfig(),
	}
}

// Flags returns the flag set for the Mithras service.
func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ExitOnError)

	cfg := DefaultConfig()
	f.Int("http_port", cfg.HTTPPort, "HTTP port to listen on")
	f.AddFlagSet(httpkit.Flags())
	f.AddFlagSet(logger.Flags())
	f.AddFlagSet(telemetry.TracingFlags())
	f.AddFlagSet(telemetry.MetricsFlags())
	f.AddFlagSet(db.Flags())
	f.AddFlagSet(ratelimit.Flags())

	return f
}

func LoadConfig(v *viper.Viper) (Config, error) {
	cfg := DefaultConfig()

	err := v.Unmarshal(&cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.TextUnmarshallerHookFunc(),
		),
	))

	return cfg, err
}
