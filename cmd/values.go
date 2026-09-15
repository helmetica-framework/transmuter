package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/helmetica-framework/transmuter/pkg/values"
)

func init() {
	valuesCmd.Flags().String("path", ".", "path to the reagent chart directory")
	valuesCmd.Flags().String("name", "", "claim name the expressions see; empty uses the chart's name")
	valuesCmd.Flags().String("namespace", "default", "claim namespace the expressions see")
	valuesCmd.Flags().String("output", values.DefaultOutput, "file to write, relative to --path")
	valuesCmd.Flags().StringArrayP("values", "f", nil, "value file layered on the chart defaults, repeatable, merged the way helm -f merges")
	valuesCmd.Flags().Bool("check", false, "compare instead of write, and fail when the file is stale")

	RootCmd.AddCommand(valuesCmd)
}

var valuesCmd = &cobra.Command{
	Use:   "values",
	Short: "Generates the values a reagent's cel: expressions compute",
	Long:  "Generates the values a reagent's cel: expressions compute, so helm-unittest can render the chart without them. With --check it compares instead of writing and fails when the file is stale. Flags can also be set via environment variables (e.g. TRANSMUTER_NAMESPACE).",
	Args:  cobra.NoArgs,
	RunE:  runValues,
}

func runValues(_ *cobra.Command, _ []string) error {
	return values.Values(values.Options{
		Path:       viper.GetString("path"),
		Name:       viper.GetString("name"),
		Namespace:  viper.GetString("namespace"),
		Output:     viper.GetString("output"),
		ValueFiles: viper.GetStringSlice("values"),
		Check:      viper.GetBool("check"),
	})
}
