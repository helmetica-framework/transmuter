package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/helmetica-framework/transmuter/pkg/validate"
)

func init() {
	validateCmd.Flags().String("path", ".", "path to the reagent chart directory")
	validateCmd.Flags().String("published-url", "", "OCI URL of the published reagent (e.g. oci://ghcr.io/helmetica-framework/myreagent); enables the CRD breakage check")

	RootCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates a reagent and checks its CRD for breaking changes",
	Long:  "Validates a reagent and checks its generated CRD for breaking changes against the latest published version. Flags can also be set via environment variables (e.g. TRANSMUTER_PUBLISHED_URL).",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		viper.SetEnvPrefix("transmuter")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.AutomaticEnv()
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: runValidate,
}

func runValidate(_ *cobra.Command, _ []string) error {
	return validate.Validate(viper.GetString("path"), viper.GetString("published-url"))
}
