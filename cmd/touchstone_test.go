package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTouchstoneAddCommandRegistered(t *testing.T) {
	var touchstoneCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "touchstone" {
			touchstoneCmd = c
			break
		}
	}
	require.NotNil(t, touchstoneCmd, "touchstone command not registered on RootCmd")

	var addCmd *cobra.Command
	for _, c := range touchstoneCmd.Commands() {
		if c.Name() == "add" {
			addCmd = c
			break
		}
	}
	require.NotNil(t, addCmd, "add subcommand not registered on touchstone command")

	pathFlag := addCmd.Flags().Lookup("path")
	require.NotNil(t, pathFlag, "--path flag missing")
	assert.Equal(t, ".", pathFlag.DefValue)

	nameFlag := addCmd.Flags().Lookup("name")
	require.NotNil(t, nameFlag, "--name flag missing")
	assert.Equal(t, "", nameFlag.DefValue)
}
