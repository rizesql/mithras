// Package serve provides the serve command for the Mithras service.
package serve

import (
	"fmt"

	"github.com/rizesql/mithras/services/mithras"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Command returns the serve command for the Mithras service.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the Mithras service",
		RunE:  serve,
		Args:  cobra.NoArgs,
	}

	cmd.Flags().AddFlagSet(mithras.Flags())
	cmd.SilenceUsage = true
	return cmd
}

func serve(cmd *cobra.Command, _ []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	cfg, err := mithras.LoadConfig(viper.GetViper())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return mithras.Run(cmd.Context(), &cfg)
}
