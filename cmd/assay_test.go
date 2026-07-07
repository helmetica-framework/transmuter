package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssayCommandRegistered(t *testing.T) {
	var assayCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "assay" {
			assayCmd = c
			break
		}
	}
	require.NotNil(t, assayCmd, "assay command not registered on RootCmd")

	pathFlag := assayCmd.Flags().Lookup("path")
	require.NotNil(t, pathFlag, "--path flag missing")
	assert.Equal(t, ".", pathFlag.DefValue)

	publishedFlag := assayCmd.Flags().Lookup("published-url")
	require.NotNil(t, publishedFlag, "--published-url flag missing")
	assert.Equal(t, "", publishedFlag.DefValue)
}
