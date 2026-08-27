package transmute

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

func Test_primaMateriaFromURL(t *testing.T) {
	tests := []struct {
		name           string
		rawURL         string
		version        string
		wantName       string
		wantRepository string
		wantErr        bool
	}{
		{
			name:           "simple repository URL",
			rawURL:         "https://charts.appcat.ch/vshnpostgresql",
			version:        "0.8.0",
			wantName:       "vshnpostgresql",
			wantRepository: "https://charts.appcat.ch",
		},
		{
			name:           "trailing slash",
			rawURL:         "https://charts.appcat.ch/vshnpostgresql/",
			version:        "0.8.0",
			wantName:       "vshnpostgresql",
			wantRepository: "https://charts.appcat.ch",
		},
		{
			name:    "no path",
			rawURL:  "https://",
			version: "0.8.0",
			wantErr: true,
		},
		{
			name:    "no slash",
			rawURL:  "vshnpostgresql",
			version: "0.8.0",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep, err := primaMateriaFromURL(tt.rawURL, tt.version)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, dep.Name)
			assert.Equal(t, tt.wantRepository, dep.Repository)
			assert.Equal(t, tt.version, dep.Version)
		})
	}
}

func Test_resolveChartRef(t *testing.T) {
	staticTags := func(tags []string, err error) func(string) ([]string, error) {
		return func(string) ([]string, error) { return tags, err }
	}

	tests := []struct {
		name     string
		ref      string
		listTags func(string) ([]string, error)
		want     string
		wantErr  bool
	}{
		{
			name:     "ref with tag returned unchanged, no lookup",
			ref:      "oci://ghcr.io/helmetica-framework/ferment:0.0.1",
			listTags: staticTags(nil, fmt.Errorf("must not be called")),
			want:     "oci://ghcr.io/helmetica-framework/ferment:0.0.1",
		},
		{
			name:     "ref without tag gets first listed tag",
			ref:      "oci://ghcr.io/helmetica-framework/ferment",
			listTags: staticTags([]string{"1.2.0", "1.1.0"}, nil),
			want:     "oci://ghcr.io/helmetica-framework/ferment:1.2.0",
		},
		{
			name:     "registry host with port is not a tag",
			ref:      "oci://localhost:5000/ferment",
			listTags: staticTags([]string{"0.3.0"}, nil),
			want:     "oci://localhost:5000/ferment:0.3.0",
		},
		{
			name:     "registry host with port and tag unchanged",
			ref:      "oci://localhost:5000/ferment:0.3.0",
			listTags: staticTags(nil, fmt.Errorf("must not be called")),
			want:     "oci://localhost:5000/ferment:0.3.0",
		},
		{
			name:     "empty tag list is an error",
			ref:      "oci://ghcr.io/helmetica-framework/ferment",
			listTags: staticTags([]string{}, nil),
			wantErr:  true,
		},
		{
			name:     "tag listing failure is an error",
			ref:      "oci://ghcr.io/helmetica-framework/ferment",
			listTags: staticTags(nil, fmt.Errorf("registry unreachable")),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveChartRef(tt.ref, tt.listTags)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_reagentMetadata(t *testing.T) {
	primaMateria := &chart.Dependency{
		Name:       "vshnpostgresql",
		Version:    "0.8.0",
		Repository: "https://charts.appcat.ch",
		Alias:      "service",
	}

	t.Run("the ferment's dependencies come along", func(t *testing.T) {
		azoth := &chart.Dependency{
			Name:       "azoth",
			Version:    "0.1.0",
			Repository: "oci://ghcr.io/helmetica-framework",
		}

		got, err := reagentMetadata(&chart.Metadata{
			Name:         "ferment",
			Version:      "0.1.0",
			Dependencies: []*chart.Dependency{azoth},
		}, "myreagent", primaMateria)
		require.NoError(t, err)

		assert.Equal(t, []*chart.Dependency{azoth, primaMateria}, got.Dependencies)
	})

	t.Run("the reagent is named and versioned as its own chart", func(t *testing.T) {
		got, err := reagentMetadata(&chart.Metadata{
			Name:       "ferment",
			Version:    "0.1.0",
			Type:       "application",
			APIVersion: "v2",
		}, "myreagent", primaMateria)
		require.NoError(t, err)

		assert.Equal(t, "myreagent", got.Name)
		assert.Equal(t, "0.0.1", got.Version)
		assert.Equal(t, "application", got.Type)
		assert.Equal(t, chart.APIVersionV2, got.APIVersion)
		require.NoError(t, got.Validate())
	})

	// Both describe the ferment. An appVersion carried over would end up in the reagent's
	// version label, claiming the ferment's version for the service.
	t.Run("the ferment's description and appVersion stay behind", func(t *testing.T) {
		got, err := reagentMetadata(&chart.Metadata{
			Name:        "ferment",
			Version:     "0.1.0",
			Description: "A starter chart for helmetica transmuter charts",
			AppVersion:  "1.16.0",
		}, "myreagent", primaMateria)
		require.NoError(t, err)

		assert.Empty(t, got.Description)
		assert.Empty(t, got.AppVersion)
	})

	// The ferment is the one place a reagent's chart metadata is configured, so anything
	// else it declares is passed on rather than dropped.
	t.Run("annotations are passed on", func(t *testing.T) {
		got, err := reagentMetadata(&chart.Metadata{
			Name:        "ferment",
			Version:     "0.1.0",
			Annotations: map[string]string{"crd.bundle.helmetica.io/kind": "Instance"},
		}, "myreagent", primaMateria)
		require.NoError(t, err)

		assert.Equal(t, map[string]string{"crd.bundle.helmetica.io/kind": "Instance"}, got.Annotations)
	})

	t.Run("the ferment's own metadata is left alone", func(t *testing.T) {
		ferment := &chart.Metadata{
			Name:         "ferment",
			Version:      "0.1.0",
			Dependencies: []*chart.Dependency{{Name: "azoth", Version: "0.1.0"}},
		}

		_, err := reagentMetadata(ferment, "myreagent", primaMateria)
		require.NoError(t, err)

		assert.Equal(t, "ferment", ferment.Name)
		assert.Len(t, ferment.Dependencies, 1)
	})

	t.Run("a ferment without metadata is an error", func(t *testing.T) {
		_, err := reagentMetadata(nil, "myreagent", primaMateria)
		require.Error(t, err)
	})
}

func Test_Transmute(t *testing.T) {
	ferment := writeFerment(t)
	t.Chdir(t.TempDir())

	err := Transmute("myreagent", ferment, "https://charts.appcat.ch/vshnpostgresql", "0.8.0")
	require.NoError(t, err)

	reagent, err := loadChart("myreagent")
	require.NoError(t, err)

	assert.Equal(t, "myreagent", reagent.Name())

	var deps []string
	for _, dep := range reagent.Metadata.Dependencies {
		deps = append(deps, dep.Name)
	}
	assert.Equal(t, []string{"azoth", "vshnpostgresql"}, deps,
		"the reagent needs the ferment's library chart as well as its prima materia")

	// Helm's starter mechanism rewrites <CHARTNAME> in templates and values.yaml.
	var templates []string
	for _, file := range reagent.Templates {
		templates = append(templates, file.Name)
		assert.NotContains(t, string(file.Data), "<CHARTNAME>")
	}
	assert.Contains(t, templates, "templates/azoth.yaml")

	gitignore, err := os.ReadFile(filepath.Join("myreagent", ".gitignore"))
	require.NoError(t, err)
	assert.Contains(t, string(gitignore), "charts/")
}

// writeFerment writes a starter chart shaped like the real ferment: a library chart it
// depends on, and one template that includes it.
func writeFerment(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "templates"), 0o755))

	files := map[string]string{
		"Chart.yaml": `apiVersion: v2
name: ferment
description: A starter chart for helmetica transmuter charts
type: application
version: 0.1.0
appVersion: "1.16.0"
dependencies:
  - name: azoth
    version: 0.1.0
    repository: oci://ghcr.io/helmetica-framework
`,
		"values.yaml":           "service: {}\n",
		"templates/azoth.yaml":  `{{- include "azoth.all" . }}` + "\n",
		"templates/ritual.yaml": `{{ include "<CHARTNAME>.labels" . }}` + "\n",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}

	return dir
}
