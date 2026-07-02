package main

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

func main() {
	slog.Info("transmutation starting")

	if len(os.Args) < 5 {
		panic("go run . $name $fermentURL $primaMateria $primaMateriaVersion")
	}

	var chartDir string
	if strings.HasPrefix(os.Args[2], "oci://") {
		slog.Info("acquiring ferment")
		downloadedTo, err := downloadChart(os.Args[2])
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(downloadedTo)
		chartDir = downloadedTo
	}

	primaMateria, err := primaMateriaFromURL(os.Args[3], os.Args[4])
	if err != nil {
		panic(err)
	}

	metada := &chart.Metadata{
		Name:       os.Args[1],
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
		panic(err)
	}

	slog.Info("assaying reagent")
	rawChart, err := loader.Load(chartDir)
	if err != nil {
		panic(err)
	}

	chrt, ok := rawChart.(*chart.Chart)
	if !ok {
		panic("reagent not a valid helm chart")
	}

	err = chrt.Validate()
	if err != nil {
		panic(fmt.Sprintf("reagent metadata invalid: %s", err))
	}

	slog.Info("reagent ready")
}

func downloadChart(chartRef string) (string, error) {
	c, err := registry.NewClient()
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
