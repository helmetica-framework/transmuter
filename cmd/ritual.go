package cmd

import (
	"github.com/helmetica-framework/transmuter/pkg/ritual"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	ritualAddCmd.Flags().String("path", ".", "path to the reagent chart directory")
	ritualAddCmd.Flags().String("name", "", "name of the ritual")

	ritualCmd.AddCommand(ritualAddCmd)
	RootCmd.AddCommand(ritualCmd)
}

var ritualCmd = &cobra.Command{
	Use:   "ritual",
	Short: "Manage ritual definitions in a reagent",
}

var ritualAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a skeleton ritual Definition to a reagent",
	Long:  "Adds a skeleton ritual Definition to a reagent. Flags can also be set via environment variables (e.g. TRANSMUTER_NAME).",
	Args:  cobra.NoArgs,
	RunE:  runRitualAdd,
}

func runRitualAdd(cmd *cobra.Command, _ []string) error {
	err := requiredParams([]string{"name"}, cmd)
	if err != nil {
		return err
	}

	return ritual.Add(viper.GetString("path"), viper.GetString("name"))
}
