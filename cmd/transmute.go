package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/helmetica-framework/transmuter/pkg/transmute"
)

const defaultFermentURL = "oci://ghcr.io/helmetica-framework/ferment"

func init() {
	transmuteCmd.Flags().String("name", "", "name of the resulting reagent")
	transmuteCmd.Flags().String("ferment-url", defaultFermentURL, "URL of the ferment to use as scaffold; without a tag the latest available version is used")
	transmuteCmd.Flags().String("prima-materia-url", "", "repository URL of the prima materia chart")
	transmuteCmd.Flags().String("prima-materia-version", "", "version of the prima materia chart")

	RootCmd.AddCommand(transmuteCmd)
}

var transmuteCmd = &cobra.Command{
	Use:   "transmute",
	Short: "Transmutes a prima materia into a valid reagent",
	Long:  "Transmutes a prima materia into a valid reagent. Flags can also be set via environment variables (e.g. TRANSMUTER_FERMENT_URL).",
	Args:  cobra.NoArgs,
	RunE:  runTransmute,
}

func runTransmute(cmd *cobra.Command, _ []string) error {
	err := requiredParams([]string{"name", "prima-materia-url", "prima-materia-version"}, cmd)
	if err != nil {
		return err
	}

	return transmute.Transmute(
		viper.GetString("name"),
		viper.GetString("ferment-url"),
		viper.GetString("prima-materia-url"),
		viper.GetString("prima-materia-version"),
	)
}
