package status

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/internal/datastore"
	"github.com/rizesql/mithras/pkg/cli"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "View the status of database migrations natively",
		RunE:  status,
		Args:  cobra.NoArgs,
	}

	cmd.Flags().AddFlagSet(datastore.Flags())
	cmd.SilenceUsage = true

	return cmd
}

func status(cmd *cobra.Command, _ []string) error {
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

	status, err := datastore.Status(cmd.Context(), &cfg)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	cli.Header("Database Migration Status")
	cli.Label("Schema", cfg.DB.SchemaName)
	cli.Label("Migrations Table", cfg.DB.MigrationsTable)
	cli.Label("Current Version", fmt.Sprintf("%d", status.CurrentVersion))
	cli.Raw("")

	if status.IsUpToDate {
		cli.Success("Database is up to date")
		cli.Subtle("Applied %d of %d migrations", status.AppliedCount, status.TotalMigrations)
	} else {
		cli.Warn("Database has pending migrations")
		cli.Subtle("Applied: %d of %d migrations", status.AppliedCount, status.TotalMigrations)
		cli.Raw("")
		cli.Label("Pending", fmt.Sprintf("%d migration(s)", status.PendingCount))

		// Show pending versions (limit to first 10 for readability)
		showCount := min(status.PendingCount, 10)

		versions := make([]string, showCount)
		for i := range versions {
			versions[i] = fmt.Sprintf("%d", status.PendingVersions[i])
		}

		cli.Subtle("Versions: %s", strings.Join(versions, ", "))

		if status.PendingCount > 10 {
			cli.Subtle("... and %d more", status.PendingCount-10)
		}

		cli.Raw("")
		cli.Info("Run 'mithras datastore migrate' to apply pending migrations")
	}

	return nil
}
