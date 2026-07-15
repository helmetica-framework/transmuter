package transmute

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v4/pkg/chart/loader"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
	"helm.sh/helm/v4/pkg/registry"
)

// Transmute transmutes a prima materia into a valid reagent.
func Transmute(name, fermentURL, primaMateriaURL, primaMateriaVersion string) error {
	slog.Info("transmutation starting")

	var chartDir string
	if strings.HasPrefix(fermentURL, "oci://") {
		slog.Info("acquiring ferment")
		downloadedTo, err := downloadChart(fermentURL)
		if err != nil {
			return err
		}
		defer os.RemoveAll(downloadedTo)
		chartDir = downloadedTo
	}

	primaMateria, err := primaMateriaFromURL(primaMateriaURL, primaMateriaVersion)
	if err != nil {
		return err
	}

	metada := &chart.Metadata{
		Name:       name,
		Type:       "application",
		APIVersion: "v2",
		Version:    "0.0.1",
		Dependencies: []*chart.Dependency{
			primaMateria,
		},
	}

	slog.Info("transmuting")
	err = chartutil.CreateFrom(metada, ".", chartDir)
	if err != nil {
		return err
	}

	slog.Info("assaying reagent")
	rawChart, err := loader.Load(chartDir)
	if err != nil {
		return err
	}

	chrt, ok := rawChart.(*chart.Chart)
	if !ok {
		return fmt.Errorf("reagent not a valid helm chart")
	}

	err = chrt.Validate()
	if err != nil {
		return fmt.Errorf("reagent metadata invalid: %w", err)
	}

	slog.Info("reagent ready")
	return nil
}

func downloadChart(chartRef string) (string, error) {
	c, err := registry.NewClient()
	if err != nil {
		return "", err
	}

	// c.Tags, unlike c.Pull, rejects the oci:// scheme prefix
	listTags := func(ref string) ([]string, error) {
		return c.Tags(strings.TrimPrefix(ref, "oci://"))
	}
	chartRef, err = resolveChartRef(chartRef, listTags)
	if err != nil {
		return "", err
	}

	res, err := c.Pull(
		chartRef,
		registry.PullOptWithChart(true),
	)
	if err != nil {
		return "", err
	}

	cacheDir, err := os.MkdirTemp("", "transmuter-*")
	if err != nil {
		return "", err
	}

	sha := fmt.Sprintf("%x", sha256.Sum256([]byte(chartRef)))
	filePath := filepath.Join(cacheDir, sha+".tgz")
	if err := os.WriteFile(filePath, res.Chart.Data, 0o644); err != nil {
		return "", err
	}

	return filePath, nil
}

func primaMateriaFromURL(rawURL, version string) (*chart.Dependency, error) {
	trimmed := strings.TrimRight(rawURL, "/")
	i := strings.LastIndex(trimmed, "/")
	if i < 0 || strings.HasSuffix(trimmed, "://") {
		return nil, fmt.Errorf("cannot derive chart name from %q", rawURL)
	}
	dep := &chart.Dependency{
		Name:       trimmed[i+1:],
		Version:    version,
		Repository: trimmed[:i],
	}
	return dep, dep.Validate()
}

// resolveChartRef appends the latest available tag when ref has none.
// A tag is a ":" after the last "/" (hosts may carry a port).
func resolveChartRef(ref string, listTags func(string) ([]string, error)) (string, error) {
	index := strings.LastIndex(ref, "/")

	if strings.Contains(ref[index:], ":") {
		return ref, nil
	}

	tags, err := listTags(ref)
	if err != nil {
		return "", fmt.Errorf("listing tags for %s: %w", ref, err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found for %s", ref)
	}

	return ref + ":" + tags[0], nil
}
