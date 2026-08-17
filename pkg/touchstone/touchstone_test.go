package touchstone

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeChartDir lays out a minimal chart directory.
func writeChartDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("apiVersion: v2\nname: probe\ntype: application\nversion: 0.1.0\n"), 0o644))
	return dir
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name        string
		chartDir    func(t *testing.T) string
		testName    string
		errContains string // empty means success expected
	}{
		{
			name:     "creates test directory in chart directory",
			chartDir: writeChartDir,
			testName: "backup",
		},
		{
			name:        "missing Chart.yaml refused",
			chartDir:    func(t *testing.T) string { return t.TempDir() },
			testName:    "backup",
			errContains: "Chart.yaml",
		},
		{
			name: "existing test directory refused",
			chartDir: func(t *testing.T) string {
				dir := writeChartDir(t)
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "test", "touchstone", "backup"), 0o755))
				return dir
			},
			testName:    "backup",
			errContains: "already exists",
		},
		{
			name:        "empty name refused",
			chartDir:    writeChartDir,
			testName:    "",
			errContains: "name",
		},
		{
			name:        "non-DNS-1123 name refused",
			chartDir:    writeChartDir,
			testName:    "Backup_Test",
			errContains: "name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.chartDir(t)
			err := Add(dir, tt.testName)
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
				return
			}
			require.NoError(t, err)

			testDir := filepath.Join(dir, "test", "touchstone", tt.testName)
			for _, file := range []string{"sources.yaml", "instance.yaml"} {
				content, readErr := os.ReadFile(filepath.Join(testDir, file))
				require.NoError(t, readErr)
				assert.NotEmpty(t, content)
			}

			raw, readErr := os.ReadFile(filepath.Join(testDir, "chainsaw-test.yaml"))
			require.NoError(t, readErr)
			content := string(raw)
			assert.Contains(t, content, "kind: Test")
			assert.Contains(t, content, "name: "+tt.testName)
			assert.NotContains(t, content, "TOUCHSTONE_NAME")
			// the bindings the generated test pre-wires for its assertions
			for _, binding := range []string{"name: chart", "name: version", "name: major", "name: id", "name: kind", "name: crd", "name: instanceNs"} {
				assert.Contains(t, content, binding)
			}
			// the scripts reach the chart root from test/touchstone/<name>
			assert.Contains(t, content, "../../../Chart.yaml")

			config, readErr := os.ReadFile(filepath.Join(dir, "test", "touchstone", "chainsaw-config.yaml"))
			require.NoError(t, readErr)
			assert.NotEmpty(t, config)
		})
	}
}

// An existing chainsaw config belongs to the reagent and is never rewritten.
func TestAddKeepsExistingConfig(t *testing.T) {
	dir := writeChartDir(t)
	configPath := filepath.Join(dir, "test", "touchstone", "chainsaw-config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, os.WriteFile(configPath, []byte("hand written"), 0o644))

	require.NoError(t, Add(dir, "backup"))

	config, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "hand written", string(config))
}

// A second touchstone lands next to the first one, config untouched.
func TestAddSecondTouchstone(t *testing.T) {
	dir := writeChartDir(t)
	require.NoError(t, Add(dir, "install"))
	require.NoError(t, Add(dir, "backup"))

	for _, name := range []string{"install", "backup"} {
		_, err := os.Stat(filepath.Join(dir, "test", "touchstone", name, "chainsaw-test.yaml"))
		require.NoError(t, err)
	}
}
