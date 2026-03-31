package datastore

import (
	"github.com/spf13/cobra"

	"github.com/rizesql/mithras/cmd/mithras/datastore/migrate"
	"github.com/rizesql/mithras/cmd/mithras/datastore/status"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datastore",
		Short: "Manage the database datastore",
	}

	cmd.AddCommand(migrate.Command())
	cmd.AddCommand(status.Command())

	return cmd
}
