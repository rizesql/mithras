// Package main provides the mithras CLI application.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"

	"github.com/rizesql/mithras/cmd/mithras/serve"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func banner() {
	c := color.New(color.FgHiBlue).SprintFunc()
	title := color.New(color.FgWhite, color.Bold).SprintFunc()
	subtle := color.New(color.FgHiBlack).SprintFunc()

	fmt.Println()
	fmt.Printf("  %s    %s\n", c(" ▇▇      ▇▇ "), title("M I T H R A S"))
	fmt.Printf("  %s    %s\n", c(" ▇▇▇▇  ▇▇▇▇ "), title("Identity Provider"))
	fmt.Printf("  %s      \n", c(" ▇▇▇▇▇▇▇▇▇▇ "))
	fmt.Printf("  %s    %s\n", c(" ▇▇  ▇▇  ▇▇ "), subtle("Self-contained authentication and authorization"))
	fmt.Printf("  %s    %s\n", c("▇▇▇▇    ▇▇▇▇"), subtle("with JWS tokens, audit logging, and rate limiting."))
	fmt.Println()
}

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

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile,
		"config",
		"",
		"config file (default is `./mithras.yaml`)",
	)

	help := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		banner()
		help(cmd, args)
	})

	rootCmd.AddCommand(serve.Command())

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		log.Fatal(err)
	}
}
