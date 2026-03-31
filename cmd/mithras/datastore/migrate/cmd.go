package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/pkg/cli"
	"github.com/rizesql/mithras/pkg/logger"
	"github.com/rizesql/mithras/services/datastore"
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
	out := cli.Default()

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	cfg, err := datastore.LoadConfig(viper.GetViper())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger.Configure(cfg.Logs)
	logger.SetHandler(cli.NewSlogHandler(out))

	if cfg.DB.URI == "" {
		return fmt.Errorf("database URI not configured")
	}

	out.Header("Running Database Migrations")
	out.Label("Schema", cfg.DB.SchemaName)
	out.Label("Migrations Table", cfg.DB.MigrationsTable)
	out.Raw("")

	if err := datastore.Migrate(cmd.Context(), &cfg); err != nil {
		return err
	}

	out.Success("Migrations completed successfully")
	return nil
}
