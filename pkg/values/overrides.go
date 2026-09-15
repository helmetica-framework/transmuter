package values

import (
	"bytes"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"

	"github.com/helmetica-framework/chrysopoeia/pkg/schemagen/parser"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
)

// overrides merges the value files the way helm merges -f: later files win,
// maps merge key by key, and anything else, a list included, is replaced whole.
// Every error names the file it came from.
func overrides(files []string) (map[string]any, error) {
	merged := map[string]any{}

	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading value file: %w", err)
		}

		vals, err := loader.LoadValues(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}

		if err := rejectExpressions(nil, vals); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}

		merged = loader.MergeMaps(merged, vals)
	}

	// whole only rewrites numbers, so a map comes back a map.
	return whole(merged).(map[string]any), nil
}

// rejectExpressions fails on a cel: string anywhere in a value file. Nothing
// evaluates it: the expressions are compiled from the chart's own values.yaml,
// so this one would reach Helm as a literal string and break a render further
// away from the file that caused it.
func rejectExpressions(path []string, val any) error {
	switch v := val.(type) {
	case string:
		if parser.IsCelExpression(v) {
			return fmt.Errorf("%s: a cel: expression in a value file is never evaluated, it belongs in the chart's values.yaml", display(path))
		}

	case map[string]any:
		for _, key := range slices.Sorted(maps.Keys(v)) {
			if err := rejectExpressions(append(slices.Clone(path), key), v[key]); err != nil {
				return err
			}
		}

	case []any:
		for i, item := range v {
			if err := rejectExpressions(append(slices.Clone(path), fmt.Sprintf("[%d]", i)), item); err != nil {
				return err
			}
		}
	}

	return nil
}

// whole converts the integral float64 values sigs.k8s.io/yaml produces into
// int64, the type a claim's values have when the controller evaluates the same
// expressions. Without it `cel: values.replicas / 2` fails with `no such
// overload` on a value the cluster divides as integers.
func whole(v any) any {
	switch v := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[key] = whole(val)
		}
		return out

	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = whole(item)
		}
		return out

	case float64:
		if v != math.Trunc(v) || v < math.MinInt64 || v > math.MaxInt64 {
			return v
		}
		return int64(v)

	default:
		return v
	}
}
