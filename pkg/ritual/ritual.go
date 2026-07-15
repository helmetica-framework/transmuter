// Package ritual scaffolds ritual Definition manifests into reagent charts.
package ritual

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

// skeleton is the ritual Definition written by Add; both %s verbs receive
// the ritual name.
//
//go:embed files/skeleton.yaml
var skeleton string

// Add writes a skeleton ritual Definition to
// <chartPath>/templates/rituals/<name>.yaml. It refuses to overwrite an
// existing file and requires chartPath to contain a Chart.yaml.
func Add(chartPath, name string) error {
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return fmt.Errorf("invalid ritual name %q: %s", name, strings.Join(errs, "; "))
	}

	if _, err := os.Stat(filepath.Join(chartPath, "Chart.yaml")); err != nil {
		return fmt.Errorf("not a chart folder: %w", err)
	}

	ritualsDir := filepath.Join(chartPath, "templates", "rituals")
	ritualFile := filepath.Join(ritualsDir, name+".yaml")

	if err := os.MkdirAll(ritualsDir, 0o755); err != nil {
		return fmt.Errorf("creating rituals folder: %w", err)
	}

	if _, err := os.Stat(ritualFile); err == nil {
		return fmt.Errorf("%s already exists", ritualFile)
	}

	if err := os.WriteFile(ritualFile, fmt.Appendf(nil, skeleton, name, name), 0o644); err != nil {
		return fmt.Errorf("writing ritual: %w", err)
	}

	slog.Info("ritual skeleton written", "path", ritualFile)
	return nil
}
