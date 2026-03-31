package datastore

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/logger"
)

type Config struct {
	Logs logger.Config `mapstructure:"logs"`
	DB   db.Config     `mapstructure:"db"`
}

func DefaultConfig() Config {
	return Config{
		Logs: logger.DefaultConfig(),
		DB:   db.DefaultConfig(),
	}
}

func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ExitOnError)
	f.AddFlagSet(logger.Flags())
	f.AddFlagSet(db.Flags())

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
