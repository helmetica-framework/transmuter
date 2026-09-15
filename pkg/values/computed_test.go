package values

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_computed(t *testing.T) {
	tests := []struct {
		name        string
		values      map[string]any
		pvalues     map[string]any
		want        map[string]any
		errContains string // empty means success expected
	}{
		{
			name:    "chart without expressions yields an empty map",
			values:  map[string]any{"replicas": int64(1), "name": "probe"},
			pvalues: map[string]any{"replicas": int64(1), "name": "probe"},
			want:    map[string]any{},
		},
		{
			name: "only the expression paths are kept",
			values: map[string]any{
				"service": map[string]any{"replicas": int64(3)},
				"sub": map[string]any{
					"replicas":         "cel: values.service.replicas",
					"fullnameOverride": "redis",
				},
			},
			pvalues: map[string]any{
				"service": map[string]any{"replicas": int64(3)},
				"sub": map[string]any{
					"replicas":         int64(3),
					"fullnameOverride": "redis",
				},
			},
			want: map[string]any{"sub": map[string]any{"replicas": int64(3)}},
		},
		{
			name: "several expressions under one parent keep their full paths",
			values: map[string]any{
				"sub": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{
							"cpu":    "cel: '250m'",
							"memory": "cel: '1Gi'",
						},
					},
				},
			},
			pvalues: map[string]any{
				"sub": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{"cpu": "250m", "memory": "1Gi"},
					},
				},
			},
			want: map[string]any{
				"sub": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{"cpu": "250m", "memory": "1Gi"},
					},
				},
			},
		},
		{
			name: "hint blocks are skipped",
			values: map[string]any{
				"#mode":   map[string]any{"default": "cel: values.storage"},
				"storage": "10Gi",
			},
			pvalues: map[string]any{
				"#mode":   map[string]any{"default": "cel: values.storage"},
				"storage": "10Gi",
			},
			want: map[string]any{},
		},
		{
			name: "a list is never walked into",
			values: map[string]any{
				"volumes": []any{map[string]any{"name": "cel: values.x"}},
				"x":       "data",
			},
			pvalues: map[string]any{
				"volumes": []any{map[string]any{"name": "cel: values.x"}},
				"x":       "data",
			},
			want: map[string]any{},
		},
		{
			name:        "expression path missing from the computed values",
			values:      map[string]any{"sub": map[string]any{"size": "cel: values.storage"}},
			pvalues:     map[string]any{"sub": map[string]any{}},
			errContains: "sub.size",
		},
		{
			name:        "expression path blocked by a scalar in the computed values",
			values:      map[string]any{"sub": map[string]any{"size": "cel: values.storage"}},
			pvalues:     map[string]any{"sub": "not a map"},
			errContains: "sub.size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computed(tt.values, tt.pvalues)
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
