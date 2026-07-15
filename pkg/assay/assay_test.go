package assay

import (
	"maps"
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

// writeChartFiles lays out a chart directory from a map of relative path to
// content. Directories are created as needed.
func writeChartFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return dir
}

const validRitual = `apiVersion: rituals.helmetica.io/v1
kind: Definition
metadata:
  name: restart
spec:
  description: Restart the application
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: restart
              image: kubectl
          restartPolicy: Never
`

func TestAssayRituals(t *testing.T) {
	base := map[string]string{
		"Chart.yaml":  validChartYaml,
		"values.yaml": "backup:\n  retention: 6\n",
	}
	withTemplate := func(name, content string) map[string]string {
		files := map[string]string{}
		maps.Copy(files, base)
		files[name] = content
		return files
	}

	tests := []struct {
		name        string
		files       map[string]string
		errContains string // empty means success expected
	}{
		{
			name:  "valid ritual passes",
			files: withTemplate("templates/rituals/restart.yaml", validRitual),
		},
		{
			name: "templated ritual renders and passes",
			files: withTemplate("templates/rituals/restart.yaml",
				"apiVersion: rituals.helmetica.io/v1\nkind: Definition\nmetadata:\n  name: {{ .Chart.Name }}-restart\nspec:\n  jobTemplate:\n    spec: {}\n"),
		},
		{
			name: "non-ritual documents ignored",
			files: withTemplate("templates/cm.yaml",
				"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm\ndata: {}\n"),
		},
		{
			name: "missing metadata.name fails",
			files: withTemplate("templates/rituals/broken.yaml",
				"apiVersion: rituals.helmetica.io/v1\nkind: Definition\nmetadata: {}\nspec:\n  jobTemplate:\n    spec: {}\n"),
			errContains: "metadata.name",
		},
		{
			name: "missing spec.jobTemplate fails",
			files: withTemplate("templates/rituals/broken.yaml",
				"apiVersion: rituals.helmetica.io/v1\nkind: Definition\nmetadata:\n  name: broken\nspec:\n  description: no job\n"),
			errContains: "jobTemplate",
		},
		{
			name: "broken template fails render",
			files: withTemplate("templates/rituals/broken.yaml",
				"metadata:\n  name: {{ .Values.doesnot.exist }}\n"),
			errContains: "render",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Assay(writeChartFiles(t, tt.files), "")
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
				return
			}
			require.NoError(t, err)
		})
	}
}

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
