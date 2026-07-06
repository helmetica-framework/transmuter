package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommandRegistered(t *testing.T) {
	var validateCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "validate" {
			validateCmd = c
			break
		}
	}
	require.NotNil(t, validateCmd, "validate command not registered on RootCmd")

	pathFlag := validateCmd.Flags().Lookup("path")
	require.NotNil(t, pathFlag, "--path flag missing")
	assert.Equal(t, ".", pathFlag.DefValue)

	publishedFlag := validateCmd.Flags().Lookup("published-url")
	require.NotNil(t, publishedFlag, "--published-url flag missing")
	assert.Equal(t, "", publishedFlag.DefValue)
}
