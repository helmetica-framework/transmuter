package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/helmetica-framework/transmuter/pkg/transmute"
)

func init() {
	transmuteCmd.Flags().String("name", "", "name of the resulting reagent")
	transmuteCmd.Flags().String("ferment-url", "", "URL of the ferment to use as scaffold (e.g. oci://ghcr.io/helmetica-framework/ferment:0.0.1)")
	transmuteCmd.Flags().String("prima-materia-url", "", "repository URL of the prima materia chart")
	transmuteCmd.Flags().String("prima-materia-version", "", "version of the prima materia chart")

	RootCmd.AddCommand(transmuteCmd)
}

var transmuteCmd = &cobra.Command{
	Use:   "transmute",
	Short: "Transmutes a prima materia into a valid reagent",
	Long:  "Transmutes a prima materia into a valid reagent. Flags can also be set via environment variables (e.g. TRANSMUTER_FERMENT_URL).",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		viper.SetEnvPrefix("transmuter")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.AutomaticEnv()
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: runTransmute,
}

func runTransmute(cmd *cobra.Command, _ []string) error {
	// required flags are checked here instead of via MarkFlagRequired so
	// values provided through environment variables count as set
	for _, key := range []string{"name", "ferment-url", "prima-materia-url", "prima-materia-version"} {
		if viper.GetString(key) == "" {
			cmd.SilenceUsage = false
			return fmt.Errorf("required flag --%s (or env %s) not set", key, "TRANSMUTER_"+strings.ToUpper(strings.ReplaceAll(key, "-", "_")))
		}
	}

	return transmute.Transmute(
		viper.GetString("name"),
		viper.GetString("ferment-url"),
		viper.GetString("prima-materia-url"),
		viper.GetString("prima-materia-version"),
	)
}
