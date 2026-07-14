package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRitualAddCommandRegistered(t *testing.T) {
	var ritualCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "ritual" {
			ritualCmd = c
			break
		}
	}
	require.NotNil(t, ritualCmd, "ritual command not registered on RootCmd")

	var addCmd *cobra.Command
	for _, c := range ritualCmd.Commands() {
		if c.Name() == "add" {
			addCmd = c
			break
		}
	}
	require.NotNil(t, addCmd, "add subcommand not registered on ritual command")

	pathFlag := addCmd.Flags().Lookup("path")
	require.NotNil(t, pathFlag, "--path flag missing")
	assert.Equal(t, ".", pathFlag.DefValue)

	nameFlag := addCmd.Flags().Lookup("name")
	require.NotNil(t, nameFlag, "--name flag missing")
	assert.Equal(t, "", nameFlag.DefValue)
}
