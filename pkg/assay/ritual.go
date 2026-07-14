package assay

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v4/pkg/chart/common"
	"helm.sh/helm/v4/pkg/chart/common/util"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/engine"
)

const (
	ritualAPIVersion = "rituals.helmetica.io/v1"
	ritualKind       = "Definition"
)

// ritualDoc is the subset of a rendered manifest needed for structural checks.
type ritualDoc struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec map[string]any `yaml:"spec"`
}

// validateRituals renders the chart with its default values and structurally
// checks every ritual Definition found in the rendered output.
func validateRituals(chrt *chart.Chart) error {
	vals, err := util.ToRenderValues(chrt, chrt.Values, common.ReleaseOptions{Name: "assay", Namespace: "default", Revision: 1, IsInstall: true}, common.DefaultCapabilities)
	if err != nil {
		return fmt.Errorf("rendering reagent values: %w", err)
	}

	rendered, err := engine.Render(chrt, vals)
	if err != nil {
		return fmt.Errorf("rendering reagent: %w", err)
	}

	for fileName, content := range rendered {
		if path.Base(fileName) == "NOTES.txt" {
			continue
		}

		dec := yaml.NewDecoder(strings.NewReader(content))
		for {
			var doc ritualDoc
			if err := dec.Decode(&doc); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return fmt.Errorf("parsing rendered %s: %w", fileName, err)
			}

			if doc.APIVersion != ritualAPIVersion || doc.Kind != ritualKind {
				continue
			}

			if doc.Metadata.Name == "" {
				return fmt.Errorf("ritual in %s: metadata.name missing", fileName)
			}

			if doc.Spec["jobTemplate"] == nil {
				return fmt.Errorf("ritual %q in %s: spec.jobTemplate missing", doc.Metadata.Name, fileName)
			}

			slog.Info("ritual valid", "name", doc.Metadata.Name, "file", fileName)
		}
	}

	return nil
}
