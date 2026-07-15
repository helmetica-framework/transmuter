package transmute

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
