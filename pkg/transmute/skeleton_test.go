package transmute

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_writeDocsSkeleton(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeDocsSkeleton(dir, "my-cache"))

	antora, err := os.ReadFile(filepath.Join(dir, "docs", "antora.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(antora), "name: reagent-my-cache\n")
	assert.Contains(t, string(antora), "title: My Cache\n")
	assert.Contains(t, string(antora), "version: ~\n", "rosarium only accepts versionless components")

	for _, page := range []string{
		"index.adoc",
		"tutorials/first-instance.adoc",
		"how-tos/connect.adoc",
		"how-tos/restore.adoc",
		"how-tos/expose.adoc",
		"how-tos/change-plan.adoc",
		"references/release-notes.adoc",
	} {
		assert.FileExists(t, filepath.Join(dir, "docs", "modules", "ROOT", "pages", page))
	}
	index, err := os.ReadFile(filepath.Join(dir, "docs", "modules", "ROOT", "pages", "index.adoc"))
	require.NoError(t, err)
	assert.Contains(t, string(index), "= My Cache")
	assert.NotContains(t, string(index), "{{")

	helmignore, err := os.ReadFile(filepath.Join(dir, ".helmignore"))
	require.NoError(t, err)
	assert.Contains(t, string(helmignore), "docs/", "docs must never be packaged into the chart")
}

func Test_writeDocsSkeleton_keepsExistingHelmignore(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".helmignore"), []byte(".git/\n"), 0o644))
	require.NoError(t, writeDocsSkeleton(dir, "redis"))
	helmignore, err := os.ReadFile(filepath.Join(dir, ".helmignore"))
	require.NoError(t, err)
	assert.Equal(t, ".git/\n# Docs are published on docs.helmetica.io, not packaged.\ndocs/\n", string(helmignore))
}

func Test_writeDocsSkeleton_reagentPrefixNotDoubled(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeDocsSkeleton(dir, "reagent-redis"))
	antora, err := os.ReadFile(filepath.Join(dir, "docs", "antora.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(antora), "name: reagent-redis\n")
	assert.Contains(t, string(antora), "title: Redis\n")
}

func Test_writeDocsSkeleton_existingDocsUntouched(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "mine.adoc"), []byte("mine"), 0o644))
	err := writeDocsSkeleton(dir, "redis")
	require.ErrorContains(t, err, "docs/ already exists")
	_, err = os.Stat(filepath.Join(dir, "docs", "antora.yml"))
	assert.True(t, os.IsNotExist(err))
}
