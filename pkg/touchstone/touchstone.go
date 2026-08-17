// Package touchstone scaffolds chainsaw tests into reagent charts.
package touchstone

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

// namePlaceholder is replaced by the touchstone name in test. A placeholder
// instead of a %s verb: the test carries awk format strings of its own.
const namePlaceholder = "TOUCHSTONE_NAME"

var (
	// config is the chainsaw Configuration shared by every touchstone of a
	// reagent. It is only written when the reagent has none yet.
	//go:embed files/chainsaw-config.yaml
	config string

	// test is the chainsaw Test: it publishes the chart, lets chrysopoeia
	// turn it into a CRD, claims an instance and leaves the bindings ready
	// for the assertions the maintainer adds.
	//go:embed files/chainsaw-test.yaml
	test string

	// sources is the flux/chrysopoeia source trio the test applies.
	//go:embed files/sources.yaml
	sources string

	// instance is the claim the test applies once the CRD exists.
	//go:embed files/instance.yaml
	instance string
)

// Add writes a skeleton chainsaw test to
// <chartPath>/test/touchstone/<name>/ along with the sources and instance
// manifests it applies. The shared chainsaw-config.yaml is written only when
// the reagent has none. It refuses to overwrite an existing touchstone and
// requires chartPath to contain a Chart.yaml.
func Add(chartPath, name string) error {
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return fmt.Errorf("invalid touchstone name %q: %s", name, strings.Join(errs, "; "))
	}

	if _, err := os.Stat(filepath.Join(chartPath, "Chart.yaml")); err != nil {
		return fmt.Errorf("not a chart folder: %w", err)
	}

	touchstoneDir := filepath.Join(chartPath, "test", "touchstone")
	testDir := filepath.Join(touchstoneDir, name)

	if _, err := os.Stat(testDir); err == nil {
		return fmt.Errorf("%s already exists", testDir)
	}

	if err := os.MkdirAll(testDir, 0o755); err != nil {
		return fmt.Errorf("creating touchstone folder: %w", err)
	}

	// The reagent owns its chainsaw configuration once it has one.
	configFile := filepath.Join(touchstoneDir, "chainsaw-config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := os.WriteFile(configFile, []byte(config), 0o644); err != nil {
			return fmt.Errorf("writing chainsaw config: %w", err)
		}
		slog.Info("chainsaw config written", "path", configFile)
	}

	files := map[string]string{
		"chainsaw-test.yaml": strings.ReplaceAll(test, namePlaceholder, name),
		"sources.yaml":       sources,
		"instance.yaml":      instance,
	}
	for fileName, content := range files {
		path := filepath.Join(testDir, fileName)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", fileName, err)
		}
	}

	slog.Info("touchstone skeleton written", "path", testDir)
	return nil
}
