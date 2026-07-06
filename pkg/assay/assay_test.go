package assay

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeChart lays out a minimal chart directory for failure-case tests.
// Empty content strings mean "do not write this file".
func writeChart(t *testing.T, chartYaml, valuesYaml string) string {
	t.Helper()
	dir := t.TempDir()
	if chartYaml != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chartYaml), 0o644))
	}
	if valuesYaml != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(valuesYaml), 0o644))
	}
	return dir
}

const validChartYaml = `apiVersion: v2
name: probe
type: application
version: 0.1.0
`

func TestAssay(t *testing.T) {
	tests := []struct {
		name        string
		path        func(t *testing.T) string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid fixture chart passes without published URL",
			path: func(t *testing.T) string { return "./testdata" },
		},
		{
			name: "missing Chart.yaml fails",
			path: func(t *testing.T) string {
				return writeChart(t, "", "backup:\n  retention: 6\n")
			},
			wantErr: true,
		},
		{
			name: "loose semver version fails strict check",
			path: func(t *testing.T) string {
				return writeChart(t, "apiVersion: v2\nname: probe\ntype: application\nversion: \"1.0\"\n", "backup:\n  retention: 6\n")
			},
			wantErr: true,
		},
		{
			name: "broken values.yaml fails",
			path: func(t *testing.T) string {
				return writeChart(t, validChartYaml, "backup: [unclosed\n")
			},
			wantErr: true,
		},
		{
			name: "missing values.yaml fails CRD generation",
			path: func(t *testing.T) string {
				return writeChart(t, validChartYaml, "")
			},
			wantErr:     true,
			errContains: "values.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Assay(tt.path(t), "")
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_latestTag(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		major   uint64
		want    string
		wantErr bool
	}{
		{
			name: "highest semver wins, not lexical order",
			tags: []string{"0.1.0", "0.10.0", "0.2.0"},
			want: "0.10.0",
		},
		{
			name: "non-semver tags ignored",
			tags: []string{"latest", "0.1.0", "not-a-version"},
			want: "0.1.0",
		},
		{
			name:  "other majors excluded",
			tags:  []string{"0.9.0", "1.0.0", "1.1.0", "2.0.0"},
			major: 1,
			want:  "1.1.0",
		},
		{
			name:    "no tags of requested major",
			tags:    []string{"0.9.0", "0.10.0"},
			major:   1,
			wantErr: true,
		},
		{
			name:    "no tags",
			tags:    []string{},
			wantErr: true,
		},
		{
			name:    "only non-semver tags",
			tags:    []string{"latest", "stable"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := latestTag(tt.tags, tt.major)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
