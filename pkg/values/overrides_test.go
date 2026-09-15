package values

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFiles lays out value files in a fresh directory and returns it.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}
	return dir
}

func Test_overrides(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		order       []string // the file names, in the order they are passed
		want        map[string]any
		errContains string // empty means success expected
	}{
		{
			name: "no files yields an empty map",
			want: map[string]any{},
		},
		{
			name:  "a single file is read as written",
			files: map[string]string{"a.yaml": "service:\n  size: 10Gi\n"},
			order: []string{"a.yaml"},
			want:  map[string]any{"service": map[string]any{"size": "10Gi"}},
		},
		{
			name: "the later file wins and maps merge key by key",
			files: map[string]string{
				"a.yaml": "service:\n  replicas: 1\n  size: 10Gi\n",
				"b.yaml": "service:\n  replicas: 3\n",
			},
			order: []string{"a.yaml", "b.yaml"},
			want: map[string]any{"service": map[string]any{
				"replicas": int64(3),
				"size":     "10Gi",
			}},
		},
		{
			name: "a list is replaced whole, not merged",
			files: map[string]string{
				"a.yaml": "ports:\n  - 80\n  - 443\n",
				"b.yaml": "ports:\n  - 8080\n",
			},
			order: []string{"a.yaml", "b.yaml"},
			want:  map[string]any{"ports": []any{int64(8080)}},
		},
		{
			name:  "integral numbers become int64, the type a claim carries in the cluster",
			files: map[string]string{"a.yaml": "whole: 3\nfraction: 1.5\nnested:\n  list:\n    - 2\n"},
			order: []string{"a.yaml"},
			want: map[string]any{
				"whole":    int64(3),
				"fraction": 1.5,
				"nested":   map[string]any{"list": []any{int64(2)}},
			},
		},
		{
			name:        "a cel: expression in a value file is rejected",
			files:       map[string]string{"a.yaml": "sub:\n  size: \"cel: values.service.size\"\n"},
			order:       []string{"a.yaml"},
			errContains: "sub.size",
		},
		{
			name:        "a missing file is an error naming it",
			order:       []string{"nope.yaml"},
			errContains: "nope.yaml",
		},
		{
			name:        "a file that is not a mapping is an error naming it",
			files:       map[string]string{"a.yaml": "- one\n- two\n"},
			order:       []string{"a.yaml"},
			errContains: "a.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeFiles(t, tt.files)

			paths := make([]string, 0, len(tt.order))
			for _, name := range tt.order {
				paths = append(paths, filepath.Join(dir, name))
			}

			got, err := overrides(paths)
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
