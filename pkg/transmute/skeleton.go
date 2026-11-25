package transmute

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// skeleton is the docs/ tree every new reagent starts with. rosarium renders
// it into docs.helmetica.io and adds the generated Parameters page.
//
//go:embed all:skeleton
var skeleton embed.FS

const helmignoreDocs = "# Docs are published on docs.helmetica.io, not packaged.\ndocs/\n"

type skeletonData struct {
	Name  string // Antora component name, the reagent's repo name
	Title string
}

// writeDocsSkeleton writes the reagent docs skeleton into reagentDir and keeps
// docs/ out of the packaged chart.
func writeDocsSkeleton(reagentDir, name string) error {
	if _, err := os.Stat(filepath.Join(reagentDir, "docs")); err == nil {
		return fmt.Errorf("%s: docs/ already exists", reagentDir)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	base := strings.TrimPrefix(name, "reagent-")
	data := skeletonData{Name: "reagent-" + base, Title: titleCase(base)}

	err := fs.WalkDir(skeleton, "skeleton", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(reagentDir, strings.TrimPrefix(path, "skeleton"))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := skeleton.ReadFile(path)
		if err != nil {
			return err
		}
		tmpl, err := template.New(path).Option("missingkey=error").Parse(string(raw))
		if err != nil {
			return err
		}
		var out bytes.Buffer
		if err := tmpl.Execute(&out, data); err != nil {
			return err
		}
		return os.WriteFile(target, out.Bytes(), 0o644)
	})
	if err != nil {
		return err
	}

	helmignore := filepath.Join(reagentDir, ".helmignore")
	existing, err := os.ReadFile(helmignore)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if bytes.Contains(existing, []byte("\ndocs/")) || bytes.HasPrefix(existing, []byte("docs/")) {
		return nil
	}
	if len(existing) > 0 && !bytes.HasSuffix(existing, []byte("\n")) {
		existing = append(existing, '\n')
	}
	return os.WriteFile(helmignore, append(existing, helmignoreDocs...), 0o644)
}

func titleCase(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}
