package transmute

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"helm.sh/helm/v4/pkg/chart/loader"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
	"helm.sh/helm/v4/pkg/registry"
)

// Transmute transmutes a prima materia into a valid reagent.
func Transmute(name, fermentURL, primaMateriaURL, primaMateriaVersion string) error {
	slog.Info("transmutation starting")

	chartDir := fermentURL
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

	ferment, err := loadChart(chartDir)
	if err != nil {
		return fmt.Errorf("ferment not a valid helm chart: %w", err)
	}

	metadata, err := reagentMetadata(ferment.Metadata, name, primaMateria)
	if err != nil {
		return err
	}

	slog.Info("transmuting")
	err = chartutil.CreateFrom(metadata, ".", chartDir)
	if err != nil {
		return err
	}

	// the starter chart can't carry a .gitignore: helm strips dotfiles on load
	gitignore := []byte("*.tgz\ncharts/\n")
	if err := os.WriteFile(filepath.Join(name, ".gitignore"), gitignore, 0o644); err != nil {
		return err
	}

	slog.Info("assaying reagent")
	reagent, err := loadChart(name)
	if err != nil {
		return fmt.Errorf("reagent not a valid helm chart: %w", err)
	}

	err = reagent.Validate()
	if err != nil {
		return fmt.Errorf("reagent metadata invalid: %w", err)
	}

	slog.Info("reagent ready")
	return nil
}

// reagentMetadata is the ferment's metadata under the reagent's name, with the prima
// materia added to its dependencies. Everything else the ferment declares carries over,
// so a library chart it depends on reaches the reagent without the transmuter knowing
// what it is called. Description and appVersion are the ferment's own and are dropped.
func reagentMetadata(ferment *chart.Metadata, name string, primaMateria *chart.Dependency) (*chart.Metadata, error) {
	if ferment == nil {
		return nil, fmt.Errorf("ferment has no metadata")
	}

	metadata := *ferment
	metadata.Name = name
	metadata.Version = "0.0.1"
	metadata.Type = "application"
	metadata.APIVersion = chart.APIVersionV2
	metadata.Description = ""
	metadata.AppVersion = ""
	metadata.Dependencies = append(slices.Clone(ferment.Dependencies), primaMateria)

	return &metadata, nil
}

func loadChart(path string) (*chart.Chart, error) {
	raw, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

	chrt, ok := raw.(*chart.Chart)
	if !ok {
		return nil, fmt.Errorf("%s is not a v2 chart", path)
	}

	return chrt, nil
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
		Alias:      "service",
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
