package transmute

import (
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
