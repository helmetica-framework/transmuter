package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func requiredParams(params []string, cmd *cobra.Command) error {
	// required flags are checked here instead of via MarkFlagRequired so
	// values provided through environment variables count as set
	for _, key := range params {
		if viper.GetString(key) == "" {
			cmd.SilenceUsage = false
			return fmt.Errorf("required flag --%s (or env %s) not set", key, "TRANSMUTER_"+strings.ToUpper(strings.ReplaceAll(key, "-", "_")))
		}
	}

	return nil
}
