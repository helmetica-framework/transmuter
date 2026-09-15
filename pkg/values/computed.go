// Package values generates the values chrysopoeia computes from a reagent's
// cel: expressions, so helm-unittest can render the chart without them.
package values

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/helmetica-framework/chrysopoeia/pkg/schemagen/parser"
)

// computed returns the subset of pvalues that celvalues computed: the value at
// every path in values that held a cel: expression, and nothing else.
func computed(values, pvalues map[string]any) (map[string]any, error) {
	out := map[string]any{}
	if err := collect(nil, values, pvalues, out); err != nil {
		return nil, err
	}
	return out, nil
}

// collect walks src the way celvalues.scan does, and copies the value at the
// path of every expression it finds out of pvalues and into out. Keys are
// walked in sorted order so that a chart wn.
func collect(path []string, src, pvalues, out map[string]any) error {
	for _, key := range slices.Sorted(maps.Keys(src)) {
		if strings.HasPrefix(key, "#") {
			// hint blocks carry schema metadata rather than values
			continue
		}

		keyPath := append(slices.Clone(path), key)

		switch v := src[key].(type) {
		case string:
			if !parser.IsCelExpression(v) {
				continue
			}
			val, err := lookup(pvalues, keyPath)
			if err != nil {
				return err
			}
			set(out, keyPath, val)

		case map[string]any:
			if err := collect(keyPath, v, pvalues, out); err != nil {
				return err
			}

			// Anything else, a list included, cannot hold an expression:
			// celvalues rejects one inside a list rather than evaluating it.
		}
	}

	return nil
}

// lookup reads the value at path out of the pre-processed values. A path that
// is absent, or blocked by something that is not a map on the way down, means
// this walk and celvalues disagree about what an expression is. Writing a
// plausible value instead would produce exactly the silently wrong file the
// command exists to prevent.
func lookup(pvalues map[string]any, path []string) (any, error) {
	var cur any = pvalues

	for _, key := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: the computed values hold no map at this path", display(path))
		}

		cur, ok = m[key]
		if !ok {
			return nil, fmt.Errorf("%s: missing from the computed values", display(path))
		}
	}

	return cur, nil
}

// set writes val into out at path, creating the maps along the way.
func set(out map[string]any, path []string, val any) {
	cur := out

	for _, key := range path[:len(path)-1] {
		next, ok := cur[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[key] = next
		}
		cur = next
	}

	cur[path[len(path)-1]] = val
}

// display renders a path for a human: dots between keys, no dot before a list
// index. No expression path runs through a list, but a value file can hold one.
func display(path []string) string {
	var b strings.Builder

	for i, segment := range path {
		if i > 0 && !strings.HasPrefix(segment, "[") {
			b.WriteByte('.')
		}
		b.WriteString(segment)
	}

	return b.String()
}
