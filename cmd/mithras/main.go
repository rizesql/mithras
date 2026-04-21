// Package main provides the mithras CLI application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rizesql/mithras/cmd/mithras/datastore"
	"github.com/rizesql/mithras/cmd/mithras/serve"
	"github.com/rizesql/mithras/pkg/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	var cfgFile string

	rootCmd := &cobra.Command{
		Use: "mithras",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if cfgFile != "" {
				viper.SetConfigFile(cfgFile)
			} else {
				viper.AddConfigPath(".")
				viper.SetConfigName("mithras")
			}

			viper.SetEnvPrefix("MITHRAS")
			viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
			viper.AutomaticEnv()

			if err := viper.ReadInConfig(); err != nil {
				if cfgFile != "" {
					return fmt.Errorf("failed to read specific config file: %w", err)
				}
			}

			cli.Configure(true, slog.LevelInfo)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile,
		"config",
		"",
		"config file (default is `./mithras.yaml`)",
	)

	rootCmd.PersistentFlags().AddFlagSet(cli.Default().Flags())

	help := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cli.PrintBanner()
		help(cmd, args)
	})

	rootCmd.AddCommand(datastore.Command())
	rootCmd.AddCommand(serve.Command())

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		cli.Fatal("%v", err)
	}
}
