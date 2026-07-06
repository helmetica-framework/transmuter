package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "transmuter",
	Short: "transmuter transmutes a prima materia into a valid reagent.",
	Long:  "transmuter transmutes a prima materia into a valid reagent.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	},
}

func Execute() {
	lifetimeCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := RootCmd.ExecuteContext(lifetimeCtx); err != nil {
		os.Exit(1)
	}
}
