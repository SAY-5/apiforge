// Package diff compares two specs and reports breaking-change
// findings. A change is "breaking" if a previously-valid client
// request would now fail. The classifier covers the common
// patterns: removed paths/methods, removed/required-added
// parameters, removed responses.
package diff

import (
	"sort"

	"github.com/SAY-5/apiforge/spec"
)

type ChangeKind string

const (
	ChangePathRemoved      ChangeKind = "path_removed"
	ChangeMethodRemoved    ChangeKind = "method_removed"
	ChangeParamRemoved     ChangeKind = "param_removed"
	ChangeParamRequiredAdded ChangeKind = "param_required_added"
	ChangeResponseRemoved  ChangeKind = "response_removed"
	ChangePathAdded        ChangeKind = "path_added"
	ChangeMethodAdded      ChangeKind = "method_added"
)

type Change struct {
	Kind     ChangeKind
	Path     string
	Method   string
	Detail   string
	Breaking bool
}

func Compare(old, new *spec.Spec) []Change {
	out := []Change{}
	oldOps := indexOps(old)
	newOps := indexOps(new)

	// Removed paths/methods (breaking).
	for key, op := range oldOps {
		if _, ok := newOps[key]; !ok {
			out = append(out, Change{
				Kind: ChangeMethodRemoved, Path: key.path, Method: key.method,
				Detail: "method removed", Breaking: true,
			})
			_ = op
		}
	}
	// Added paths/methods (non-breaking).
	for key := range newOps {
		if _, ok := oldOps[key]; !ok {
			out = append(out, Change{
				Kind: ChangeMethodAdded, Path: key.path, Method: key.method,
				Detail: "method added", Breaking: false,
			})
		}
	}
	// Per-method comparisons.
	for key, oldOp := range oldOps {
		newOp, ok := newOps[key]
		if !ok {
			continue
		}
		out = append(out, compareOp(key.path, key.method, oldOp, newOp)...)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		if out[i].Method != out[j].Method {
			return out[i].Method < out[j].Method
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

type opKey struct{ path, method string }

func indexOps(s *spec.Spec) map[opKey]*spec.Operation {
	m := map[opKey]*spec.Operation{}
	for _, ref := range s.Operations() {
		m[opKey{ref.Path, ref.Method}] = ref.Op
	}
	return m
}

func compareOp(path, method string, oldOp, newOp *spec.Operation) []Change {
	out := []Change{}
	oldParams := paramSet(oldOp)
	newParams := paramSet(newOp)

	for name, op := range oldParams {
		if _, ok := newParams[name]; !ok {
			// Removed param. Breaking only if it was required.
			out = append(out, Change{
				Kind: ChangeParamRemoved, Path: path, Method: method,
				Detail: "param '" + name + "' removed", Breaking: op.Required,
			})
		}
	}
	for name, np := range newParams {
		op, ok := oldParams[name]
		if !ok {
			if np.Required {
				// New required param → breaking.
				out = append(out, Change{
					Kind: ChangeParamRequiredAdded, Path: path, Method: method,
					Detail: "required param '" + name + "' added", Breaking: true,
				})
			}
			continue
		}
		// Param existed before and now: required-flag flipped to true → breaking.
		if !op.Required && np.Required {
			out = append(out, Change{
				Kind: ChangeParamRequiredAdded, Path: path, Method: method,
				Detail: "param '" + name + "' became required", Breaking: true,
			})
		}
	}

	for code := range oldOp.Responses {
		if _, ok := newOp.Responses[code]; !ok {
			out = append(out, Change{
				Kind: ChangeResponseRemoved, Path: path, Method: method,
				Detail: "response '" + code + "' removed", Breaking: code[0] == '2',
			})
		}
	}
	return out
}

func paramSet(op *spec.Operation) map[string]spec.Parameter {
	m := map[string]spec.Parameter{}
	for _, p := range op.Parameters {
		m[p.Name] = p
	}
	return m
}

// HasBreaking reports whether any change in the slice is breaking.
// CI uses this to gate the merge.
func HasBreaking(changes []Change) bool {
	for _, c := range changes {
		if c.Breaking {
			return true
		}
	}
	return false
}
