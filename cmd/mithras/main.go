// Package main provides the mithras CLI application.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/rizesql/mithras/cmd/mithras/datastore"
	"github.com/rizesql/mithras/cmd/mithras/serve"
	"github.com/rizesql/mithras/pkg/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	out := cli.Default()

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

			out.Configure()

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile,
		"config",
		"",
		"config file (default is `./mithras.yaml`)",
	)

	rootCmd.PersistentFlags().AddFlagSet(out.Flags())

	help := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		out.PrintBanner()
		help(cmd, args)
	})

	rootCmd.AddCommand(datastore.Command())
	rootCmd.AddCommand(serve.Command())

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		out.Fatal("%v", err)
	}
}
