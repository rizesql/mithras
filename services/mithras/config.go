package mithras

import (
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/tracing"

	"github.com/spf13/pflag"
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
