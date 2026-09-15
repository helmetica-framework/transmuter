package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/helmetica-framework/transmuter/pkg/values"
)

const probeChartYaml = `apiVersion: v2
name: probe
type: application
version: 0.1.0
`

// writeProbeChart lays out a minimal reagent with the given values.yaml.
func writeProbeChart(t *testing.T, valuesYaml string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(probeChartYaml), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(valuesYaml), 0o644))
	return dir
}

// resetFlags puts every subcommand flag back to its default. cobra keeps flag
// values on the command itself, so a repeated -f, or a --check from an earlier
// test, would otherwise reach the next one.
func resetFlags(t *testing.T) {
	t.Helper()
	for _, c := range RootCmd.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if slice, ok := f.Value.(pflag.SliceValue); ok {
				require.NoError(t, slice.Replace([]string{}))
			} else {
				require.NoError(t, f.Value.Set(f.DefValue))
			}
			f.Changed = false
		})
	}
}

// execute runs the root command with args, the way the binary would. Viper and
// the flag values are reset around the run, before as well as after: cobra and
// viper both keep global state, and a test that runs the command twice has to
// see the second run the way the binary would, not with the first run's flags
// still set.
func execute(t *testing.T, args ...string) error {
	t.Helper()
	viper.Reset()
	resetFlags(t)

	t.Cleanup(viper.Reset)
	t.Cleanup(func() { resetFlags(t) })
	t.Cleanup(func() { RootCmd.SetArgs(nil) })

	RootCmd.SetOut(io.Discard)
	RootCmd.SetErr(io.Discard)
	RootCmd.SetArgs(args)

	return RootCmd.Execute()
}

func TestValuesCommandRegistered(t *testing.T) {
	var valuesCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "values" {
			valuesCmd = c
			break
		}
	}
	require.NotNil(t, valuesCmd, "values command not registered on RootCmd")

	defaults := map[string]string{
		"path":      ".",
		"name":      "",
		"namespace": "default",
		"output":    values.DefaultOutput,
		"values":    "[]",
		"check":     "false",
	}
	for name, want := range defaults {
		flag := valuesCmd.Flags().Lookup(name)
		require.NotNil(t, flag, "--%s flag missing", name)
		assert.Equal(t, want, flag.DefValue)
	}

	assert.Equal(t, "bool", valuesCmd.Flags().Lookup("check").Value.Type(), "--check must be a flag, not a string that reads false")

	shorthand := valuesCmd.Flags().ShorthandLookup("f")
	require.NotNil(t, shorthand, "-f missing")
	assert.Equal(t, "values", shorthand.Name, "-f is what a helm user reaches for")
}

func TestValuesCommandValueFiles(t *testing.T) {
	dir := writeProbeChart(t, "service:\n  replicas: 1\nsub:\n  replicas: \"cel: values.service.replicas\"\n")
	override := filepath.Join(dir, "ha.yaml")
	require.NoError(t, os.WriteFile(override, []byte("service:\n  replicas: 3\n"), 0o644))

	require.NoError(t, execute(t, "values", "--path", dir, "-f", override))

	written, err := os.ReadFile(filepath.Join(dir, values.DefaultOutput))
	require.NoError(t, err)
	assert.Contains(t, string(written), "replicas: 3", "-f has to reach the expressions")
}

func TestValuesCommandCheckReadsTheValueFiles(t *testing.T) {
	dir := writeProbeChart(t, "service:\n  replicas: 1\nsub:\n  replicas: \"cel: values.service.replicas\"\n")
	override := filepath.Join(dir, "ha.yaml")
	require.NoError(t, os.WriteFile(override, []byte("service:\n  replicas: 3\n"), 0o644))
	require.NoError(t, execute(t, "values", "--path", dir, "-f", override))

	require.Error(t, execute(t, "values", "--path", dir, "--check"), "the check has to be given the same flags")
	assert.NoError(t, execute(t, "values", "--path", dir, "-f", override, "--check"))
}

func TestValuesCommandWritesTheFile(t *testing.T) {
	dir := writeProbeChart(t, "service:\n  replicas: 3\nsub:\n  ha: \"cel: values.service.replicas > 1\"\n")

	require.NoError(t, execute(t, "values", "--path", dir, "--name", "instance"))

	written, err := os.ReadFile(filepath.Join(dir, values.DefaultOutput))
	require.NoError(t, err)
	assert.Contains(t, string(written), "ha: true")
}

func TestValuesCommandCheckFailsWhenStale(t *testing.T) {
	dir := writeProbeChart(t, "service:\n  replicas: 3\nsub:\n  ha: \"cel: values.service.replicas > 1\"\n")
	require.NoError(t, execute(t, "values", "--path", dir))

	// The expression moves, the generated file does not.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"),
		[]byte("service:\n  replicas: 3\nsub:\n  ha: \"cel: values.service.replicas > 99\"\n"), 0o644))

	err := execute(t, "values", "--path", dir, "--check")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "transmuter values", "the error must say how to fix it")
}

func TestValuesCommandCheckPassesWhenFresh(t *testing.T) {
	dir := writeProbeChart(t, "service:\n  replicas: 3\nsub:\n  ha: \"cel: values.service.replicas > 1\"\n")
	require.NoError(t, execute(t, "values", "--path", dir))

	assert.NoError(t, execute(t, "values", "--path", dir, "--check"))
}
