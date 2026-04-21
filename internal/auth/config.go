package auth

import (
	"time"

	"github.com/spf13/pflag"
)

type Config struct {
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
	MaxFailedAttempts    int           `mapstructure:"max_failed_attempts"`
	LockoutDuration      time.Duration `mapstructure:"lockout_duration"`
	KEK                  []byte        `mapstructure:"kek"`
}

func DefaultConfig() Config {
	return Config{
		AccessTokenDuration:  5 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		MaxFailedAttempts:    5,
		LockoutDuration:      15 * time.Minute,
	}
}

func Flags() *pflag.FlagSet {
	f := pflag.NewFlagSet("auth", pflag.ExitOnError)

	cfg := DefaultConfig()
	f.Duration("auth.access_token_duration", cfg.AccessTokenDuration, "duration of access token")
	f.Duration("auth.refresh_token_duration", cfg.RefreshTokenDuration, "duration of refresh token")
	f.Int("auth.max_failed_attempts", cfg.MaxFailedAttempts, "max failed login attempts before lockout")
	f.Duration("auth.lockout_duration", cfg.LockoutDuration, "duration of account lockout")
	f.BytesBase64("auth.kek", nil, "key encryption key")

	return f
}
