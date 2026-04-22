// Package serve provides the serve command for the Mithras service.
package serve

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/internal/mithras"
	"github.com/rizesql/mithras/internal/mithras/config"
)

// Command returns the serve command for the Mithras service.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the Mithras service",
		RunE:  serve,
		Args:  cobra.NoArgs,
	}

	cmd.Flags().AddFlagSet(config.Flags())
	cmd.SilenceUsage = true

	return cmd
}

func serve(cmd *cobra.Command, _ []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	cfg, err := config.Load(viper.GetViper())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return mithras.Run(cmd.Context(), &cfg)
}
