package mithras

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/tracing"
)

// Config holds the configuration for the Mithras service.
type Config struct {
	HTTPPort int            `mapstructure:"http_port"`
	Server   httpkit.Config `mapstructure:"server"`
	Tracing  tracing.Config `mapstructure:"tracing"`
}

// DefaultConfig returns the default configuration for the Mithras service.
func DefaultConfig() Config {
	return Config{
		HTTPPort: 8080,
		Server:   httpkit.DefaultConfig(),
		Tracing:  tracing.DefaultConfig(),
	}
}

// Flags returns the flag set for the Mithras service.
func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ExitOnError)

	cfg := DefaultConfig()
	f.Int("http_port", cfg.HTTPPort, "HTTP port to listen on")
	f.AddFlagSet(httpkit.Flags())
	f.AddFlagSet(tracing.Flags())

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
