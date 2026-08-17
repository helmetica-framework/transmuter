package cmd

import (
	"github.com/helmetica-framework/transmuter/pkg/touchstone"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	touchstoneAddCmd.Flags().String("path", ".", "path to the reagent chart directory")
	touchstoneAddCmd.Flags().String("name", "", "name of the touchstone")

	touchstoneCmd.AddCommand(touchstoneAddCmd)
	RootCmd.AddCommand(touchstoneCmd)
}

var touchstoneCmd = &cobra.Command{
	Use:   "touchstone",
	Short: "Manage chainsaw tests in a reagent",
}

var touchstoneAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a skeleton chainsaw test to a reagent",
	Long:  "Adds a skeleton chainsaw test to a reagent. Flags can also be set via environment variables (e.g. TRANSMUTER_NAME).",
	Args:  cobra.NoArgs,
	RunE:  runTouchstoneAdd,
}

func runTouchstoneAdd(cmd *cobra.Command, _ []string) error {
	err := requiredParams([]string{"name"}, cmd)
	if err != nil {
		return err
	}

	return touchstone.Add(viper.GetString("path"), viper.GetString("name"))
}
