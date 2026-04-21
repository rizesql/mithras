package migrate

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/internal/datastore"
	"github.com/rizesql/mithras/pkg/cli"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Executes the backend database migrations required for Mithras",
		RunE:  migrate,
		Args:  cobra.NoArgs,
	}

	cmd.Flags().AddFlagSet(datastore.Flags())
	cmd.SilenceUsage = true

	return cmd
}

func migrate(cmd *cobra.Command, _ []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	cfg, err := datastore.LoadConfig(viper.GetViper())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cli.Configure(cfg.Logs.Enabled, cfg.Logs.Level)

	if cfg.DB.URI == "" {
		return errors.New("database URI not configured")
	}

	cli.Header("Running Database Migrations")
	cli.Label("Schema", cfg.DB.SchemaName)
	cli.Label("Migrations Table", cfg.DB.MigrationsTable)
	cli.Raw("")

	if err := datastore.Migrate(cmd.Context(), &cfg); err != nil {
		return err
	}

	cli.Success("Migrations completed successfully")

	return nil
}
