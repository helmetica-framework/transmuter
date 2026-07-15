// Package assay checks that a reagent chart is valid and that its
// generated CRD does not break the previously published version.
package assay

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/helmetica-framework/chrysopoeia/pkg/breakagedetection"
	"github.com/helmetica-framework/chrysopoeia/pkg/schemagen"
	"helm.sh/helm/v4/pkg/chart/loader"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/registry"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var errNoVersions = errors.New("no semver tags found")

// Assay checks the reagent chart at path: loadable, metadata valid,
// version strict semver, values.yaml CRD-generatable. With publishedURL set,
// it also compares the generated CRD against the latest published reagent
// and errors on breaking changes.
func Assay(path, publishedURL string) error {
	slog.Info("assaying reagent", "path", path)

	chrt, err := loadChart(path)
	if err != nil {
		return err
	}

	if err := chrt.Validate(); err != nil {
		return fmt.Errorf("reagent metadata invalid: %w", err)
	}

	version, err := semver.StrictNewVersion(chrt.Metadata.Version)
	if err != nil {
		return fmt.Errorf("reagent version %q not strict semver: %w", chrt.Metadata.Version, err)
	}

	crd, err := schemagen.GenerateCRD(*chrt)
	if err != nil {
		return fmt.Errorf("generating CRD from reagent: %w", err)
	}

	if err := validateRituals(chrt); err != nil {
		return err
	}

	if publishedURL != "" {
		if err := checkBreakage(crd, publishedURL, version.Major()); err != nil {
			return err
		}
	}

	slog.Info("reagent valid")
	return nil
}

func loadChart(path string) (*chart.Chart, error) {
	rawChart, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	chrt, ok := rawChart.(*chart.Chart)
	if !ok {
		return nil, fmt.Errorf("reagent not a valid helm chart")
	}
	return chrt, nil
}

// checkBreakage compares updated against the CRD generated from the latest
// published reagent of the same major version at publishedURL. Other majors
// use a different CRD group and coexist instead of updating in place, so
// there is nothing to break across a major boundary (same scoping as the
// chrysopoeia controller's version constraint).
func checkBreakage(updated apiextv1.CustomResourceDefinition, publishedURL string, major uint64) error {
	client, err := registry.NewClient()
	if err != nil {
		return err
	}

	ref := strings.TrimPrefix(publishedURL, "oci://")
	tags, err := client.Tags(ref)
	if err != nil {
		return fmt.Errorf("listing tags of %s: %w", publishedURL, err)
	}

	latest, err := latestTag(tags, major)
	if errors.Is(err, errNoVersions) {
		slog.Warn("no published reagent for this major version, skipping breakage check", "url", publishedURL, "major", major)
		return nil
	}
	if err != nil {
		return err
	}

	slog.Info("acquiring published reagent", "version", latest)
	res, err := client.Pull(ref+":"+latest, registry.PullOptWithChart(true))
	if err != nil {
		return fmt.Errorf("pulling published reagent %s:%s: %w", ref, latest, err)
	}

	rawPublished, err := loader.LoadArchive(bytes.NewReader(res.Chart.Data))
	if err != nil {
		return err
	}
	publishedChrt, ok := rawPublished.(*chart.Chart)
	if !ok {
		return fmt.Errorf("published reagent not a valid helm chart")
	}

	original, err := schemagen.GenerateCRD(*publishedChrt)
	if err != nil {
		return fmt.Errorf("generating CRD from published reagent: %w", err)
	}

	warnings, breakErrs := breakagedetection.Check(original, updated)
	for _, w := range warnings {
		slog.Warn(w)
	}
	if len(breakErrs) > 0 {
		return fmt.Errorf("breaking CRD changes: %s", strings.Join(breakErrs, "; "))
	}
	return nil
}

// latestTag returns the highest strict-semver tag of the given major version,
// ignoring non-semver tags and other majors.
func latestTag(tags []string, major uint64) (string, error) {
	var latest *semver.Version
	for _, t := range tags {
		v, err := semver.StrictNewVersion(t)
		if err != nil || v.Major() != major {
			continue
		}
		if latest == nil || v.GreaterThan(latest) {
			latest = v
		}
	}
	if latest == nil {
		return "", errNoVersions
	}
	return latest.Original(), nil
}
