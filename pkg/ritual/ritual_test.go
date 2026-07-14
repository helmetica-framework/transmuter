package ritual

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
		ritualName  string
		errContains string // empty means success expected
	}{
		{
			name:       "creates skeleton in chart directory",
			chartDir:   writeChartDir,
			ritualName: "restart",
		},
		{
			name:        "missing Chart.yaml refused",
			chartDir:    func(t *testing.T) string { return t.TempDir() },
			ritualName:  "restart",
			errContains: "Chart.yaml",
		},
		{
			name: "existing ritual file refused",
			chartDir: func(t *testing.T) string {
				dir := writeChartDir(t)
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "templates", "rituals"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "templates", "rituals", "restart.yaml"), []byte("occupied"), 0o644))
				return dir
			},
			ritualName:  "restart",
			errContains: "already exists",
		},
		{
			name:        "empty name refused",
			chartDir:    writeChartDir,
			ritualName:  "",
			errContains: "name",
		},
		{
			name:        "non-DNS-1123 name refused",
			chartDir:    writeChartDir,
			ritualName:  "Weekly_Maintenance",
			errContains: "name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.chartDir(t)
			err := Add(dir, tt.ritualName)
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
				return
			}
			require.NoError(t, err)
			content, readErr := os.ReadFile(filepath.Join(dir, "templates", "rituals", tt.ritualName+".yaml"))
			require.NoError(t, readErr)
			assert.Contains(t, string(content), "apiVersion: rituals.helmetica.io/v1")
			assert.Contains(t, string(content), "kind: Definition")
			assert.Contains(t, string(content), "name: "+tt.ritualName)
			assert.Contains(t, string(content), "jobTemplate:")
		})
	}
}
