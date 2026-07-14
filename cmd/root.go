package cmd

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "transmuter",
	Short: "transmuter transmutes a prima materia into a valid reagent.",
	Long:  "transmuter transmutes a prima materia into a valid reagent.",
	// Wires every subcommand's flags to viper so each flag is also settable
	// via a TRANSMUTER_* environment variable (dashes become underscores).
	// cmd is the actually executed subcommand, so BindPFlags binds its flags.
	// A subcommand defining its own PersistentPreRunE would shadow this one.
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		cmd.SilenceUsage = true
		viper.SetEnvPrefix("transmuter")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.AutomaticEnv()
		return viper.BindPFlags(cmd.Flags())
	},
}

func Execute() {
	lifetimeCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := RootCmd.ExecuteContext(lifetimeCtx); err != nil {
		os.Exit(1)
	}
}
