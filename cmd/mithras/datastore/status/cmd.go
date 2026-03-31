package status

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rizesql/mithras/pkg/cli"
	"github.com/rizesql/mithras/pkg/logger"
	"github.com/rizesql/mithras/services/datastore"
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

	status, err := datastore.Status(cmd.Context(), &cfg)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	out.Header("Database Migration Status")
	out.Label("Schema", cfg.DB.SchemaName)
	out.Label("Migrations Table", cfg.DB.MigrationsTable)
	out.Label("Current Version", fmt.Sprintf("%d", status.CurrentVersion))
	out.Raw("")

	if status.IsUpToDate {
		out.Success("Database is up to date")
		out.Subtle("Applied %d of %d migrations", status.AppliedCount, status.TotalMigrations)
	} else {
		out.Warn("Database has pending migrations")
		out.Subtle("Applied: %d of %d migrations", status.AppliedCount, status.TotalMigrations)
		out.Raw("")
		out.Label("Pending", fmt.Sprintf("%d migration(s)", status.PendingCount))

		// Show pending versions (limit to first 10 for readability)
		showCount := min(status.PendingCount, 10)
		versions := make([]string, showCount)
		for i := range versions {
			versions[i] = fmt.Sprintf("%d", status.PendingVersions[i])
		}
		out.Subtle("Versions: %s", strings.Join(versions, ", "))

		if status.PendingCount > 10 {
			out.Subtle("... and %d more", status.PendingCount-10)
		}

		out.Raw("")
		out.Info("Run 'mithras datastore migrate' to apply pending migrations")
	}

	return nil
}
