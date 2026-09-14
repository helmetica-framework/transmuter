package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/helmetica-framework/transmuter/pkg/mix"
)

func init() {
	mixCmd.Flags().String("path", ".", "path to the reagent chart directory")
	mixCmd.Flags().String("namespace", "default", "namespace to mix the reagent into")

	RootCmd.AddCommand(mixCmd)
}

var mixCmd = &cobra.Command{
	Use:   "mix",
	Short: "Mixes a reagent into the cluster",
	Long:  "Mixes a reagent into the cluster. It installs the reagent chart, or upgrades it if a release already exists. Flags can also be set via environment variables (e.g. TRANSMUTER_NAMESPACE).",
	Args:  cobra.NoArgs,
	RunE:  runMix,
}

func runMix(cmd *cobra.Command, _ []string) error {
	path := viper.GetString("path")
	namespace := viper.GetString("namespace")

	chrt, err := mix.LoadChart(path, namespace)
	if err != nil {
		return fmt.Errorf("cel pre-processing: %w", err)
	}

	return mix.Mix(cmd.Context(), chrt.Chart, chrt.Pvalues, namespace)
}
