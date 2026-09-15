package mix

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeChart lays out a chart directory from a map of relative path to content.
func writeChart(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return dir
}

const probeChartYaml = `apiVersion: v2
name: probe
type: application
version: 0.1.0
`

const claimValues = "who: \"cel: claim.name\"\nwhere: \"cel: claim.namespace\"\n"

func TestLoadChartAs(t *testing.T) {
	tests := []struct {
		name          string
		claimName     string
		namespace     string
		wantName      string
		wantNamespace string
	}{
		{
			name:          "explicit claim name is used",
			claimName:     "instance",
			namespace:     "prod",
			wantName:      "instance",
			wantNamespace: "prod",
		},
		{
			name:          "empty claim name falls back to the chart name",
			claimName:     "",
			namespace:     "prod",
			wantName:      "probe",
			wantNamespace: "prod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeChart(t, map[string]string{
				"Chart.yaml":  probeChartYaml,
				"values.yaml": claimValues,
			})

			loaded, err := LoadChartAs(dir, tt.claimName, tt.namespace)
			require.NoError(t, err)

			assert.Equal(t, tt.wantName, loaded.Pvalues["who"])
			assert.Equal(t, tt.wantNamespace, loaded.Pvalues["where"])
		})
	}
}

func TestLoadChartKeepsUsingTheChartName(t *testing.T) {
	dir := writeChart(t, map[string]string{
		"Chart.yaml":  probeChartYaml,
		"values.yaml": claimValues,
	})

	loaded, err := LoadChart(dir, "prod")
	require.NoError(t, err)

	assert.Equal(t, "probe", loaded.Pvalues["who"], "mix and assay must be unaffected")
	assert.Equal(t, "prod", loaded.Pvalues["where"])
}
