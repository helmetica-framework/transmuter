package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_transmuteFermentURLDefault(t *testing.T) {
	flag := transmuteCmd.Flags().Lookup("ferment-url")
	require.NotNil(t, flag)
	assert.Equal(t, defaultFermentURL, flag.DefValue)
	assert.Equal(t, "oci://ghcr.io/helmetica-framework/ferment", defaultFermentURL)
}
