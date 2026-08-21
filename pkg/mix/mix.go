// Package mix installs a reagent chart into a cluster, upgrading the release
// if one already exists.
package mix

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/helmetica-framework/chrysopoeia/pkg/celvalues"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/chart/loader"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/cli"
	"helm.sh/helm/v4/pkg/kube"
	"helm.sh/helm/v4/pkg/release"
	"helm.sh/helm/v4/pkg/storage/driver"
)

const timeout = 5 * time.Minute

// LoadedChart is a reagent chart together with its CEL pre-processed values.
type LoadedChart struct {
	Chart   *chart.Chart
	Pvalues map[string]any
}

// LoadChart loads the reagent chart at path and resolves the CEL expressions
// in its values against a claim for the given namespace.
func LoadChart(path, namespace string) (LoadedChart, error) {
	rawChart, err := loader.Load(path)
	if err != nil {
		return LoadedChart{}, fmt.Errorf("loading chart %s: %w", path, err)
	}

	chrt, ok := rawChart.(*chart.Chart)
	if !ok {
		return LoadedChart{}, fmt.Errorf("reagent not a valid helm chart")
	}

	prep, err := celvalues.NewFromChart(chrt)
	if err != nil {
		return LoadedChart{}, fmt.Errorf("preprocessing values of %s: %w", path, err)
	}

	pvalues, err := prep.Apply(chrt.Values, celvalues.Claim{
		Name:      chrt.Name(),
		Namespace: namespace,
	})

	return LoadedChart{
		Chart:   chrt,
		Pvalues: pvalues,
	}, err
}

// Mix installs chrt with values into namespace. An existing release of the
// same name is upgraded instead of installed.
func Mix(ctx context.Context, chrt *chart.Chart, values map[string]any, namespace string) error {
	settings := cli.New()

	var cfg action.Configuration
	if err := cfg.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER")); err != nil {
		return fmt.Errorf("initializing helm: %w", err)
	}

	slog.Info("mixing reagent into cluster", "reagent", chrt.Name(), "namespace", namespace)

	history := action.NewHistory(&cfg)
	history.Max = 1

	var rel release.Releaser

	_, err := history.Run(chrt.Name())
	switch {
	case errors.Is(err, driver.ErrReleaseNotFound):
		install := action.NewInstall(&cfg)
		install.ReleaseName = chrt.Name()
		install.Namespace = namespace
		install.CreateNamespace = true
		install.WaitStrategy = kube.StatusWatcherStrategy
		install.Timeout = timeout

		rel, err = install.RunWithContext(ctx, chrt, values)
	case err != nil:
		return fmt.Errorf("checking release history: %w", err)
	default:
		upgrade := action.NewUpgrade(&cfg)
		upgrade.Namespace = namespace
		upgrade.WaitStrategy = kube.StatusWatcherStrategy
		upgrade.Timeout = timeout

		rel, err = upgrade.RunWithContext(ctx, chrt.Name(), chrt, values)
	}
	if err != nil {
		return fmt.Errorf("mixing %s: %w", chrt.Name(), err)
	}

	acc, err := release.NewAccessor(rel)
	if err != nil {
		return fmt.Errorf("reading release: %w", err)
	}

	slog.Info("reagent mixed", "release", acc.Name(), "namespace", acc.Namespace(), "revision", acc.Version(), "status", acc.Status())

	return nil
}
