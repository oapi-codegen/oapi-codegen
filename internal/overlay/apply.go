package overlay

import (
	"errors"
	"fmt"
	"slices"

	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"

	"github.com/oapi-codegen/oapi-codegen/v2/internal/overlay/legacyjsonpath"
)

// LegacyOptionHint names the configuration that selects the JSONPath dialect
// of earlier oapi-codegen versions.
const LegacyOptionHint = "compatibility.overlays-use-yaml-jsonpath: true"

// Options controls how an overlay is applied.
type Options struct {
	// Strict makes Apply fail when the target of an action matches nothing.
	// Every action is still applied, and the error lists all of them.
	Strict bool
	// LegacyJSONPath reads targets in the JSONPath dialect of earlier
	// oapi-codegen versions, by translating them to RFC 9535 with package
	// legacyjsonpath.
	LegacyJSONPath bool
}

// Apply applies the overlay's actions to doc, in order, and returns the
// result. doc is modified in place where possible; the returned value is the
// document to use, as an action on $ may replace it.
//
// An update merges into each target: into a mapping, its mappings merge
// recursively, its sequences are appended to existing sequences, and every
// other value replaces the existing one; a sequence target gets the update
// appended, or its elements when the update is a sequence; any other target
// is replaced. A remove deletes each target from its parent.
func (o *Overlay) Apply(doc any, opts Options) (any, error) {
	var unmatched []string
	for i, a := range o.Actions {
		path, err := compileTarget(a.Target, opts.LegacyJSONPath)
		if err != nil {
			return nil, fmt.Errorf("overlay %s: %w", a.describe(i), err)
		}
		nodes := path.SelectLocated(doc)
		if len(nodes) == 0 {
			if opts.Strict {
				unmatched = append(unmatched, a.describe(i))
			}
			continue
		}
		paths := uniquePaths(nodes)
		switch {
		case a.Remove:
			doc, err = remove(doc, paths)
		case a.HasUpdate:
			doc, err = update(doc, paths, a.Update)
		}
		if err != nil {
			return nil, fmt.Errorf("overlay %s: %w", a.describe(i), err)
		}
	}
	if len(unmatched) > 0 {
		return nil, &UnmatchedTargetsError{Actions: unmatched}
	}
	return doc, nil
}

// compileTarget parses target as RFC 9535 JSONPath, or as the legacy dialect
// when legacy is set. Its errors suggest how to fix the target.
func compileTarget(target string, legacy bool) (*jsonpath.Path, error) {
	if legacy {
		translated, err := legacyjsonpath.Translate(target)
		if err != nil {
			msg := fmt.Sprintf("%v (%s is set, so targets are read as yaml-jsonpath)", err, LegacyOptionHint)
			if _, rfcErr := jsonpath.Parse(target); rfcErr == nil {
				msg += "; the target is valid RFC 9535 JSONPath, which is used when the option is turned off"
			}
			return nil, errors.New(msg)
		}
		path, err := jsonpath.Parse(translated)
		if err != nil {
			return nil, fmt.Errorf("target %q was translated to %q, which is not valid RFC 9535 JSONPath: %w", target, translated, err)
		}
		return path, nil
	}

	path, err := jsonpath.Parse(target)
	if err == nil {
		return path, nil
	}
	msg := fmt.Sprintf("target %q is not valid RFC 9535 JSONPath: %v", target, err)
	if translated, terr := legacyjsonpath.Translate(target); terr == nil {
		if _, perr := jsonpath.Parse(translated); perr == nil {
			msg += fmt.Sprintf("; if it is written for an earlier oapi-codegen version, the RFC 9535 equivalent is probably %q", translated)
		}
	}
	msg += "; to keep reading targets in the JSONPath dialect of earlier oapi-codegen versions, set " + LegacyOptionHint
	return nil, errors.New(msg)
}

// uniquePaths returns the distinct locations of nodes.
func uniquePaths(nodes jsonpath.LocatedNodeList) []spec.NormalizedPath {
	paths := make([]spec.NormalizedPath, 0, len(nodes))
	for _, n := range nodes.Deduplicate() {
		paths = append(paths, n.Path)
	}
	return paths
}

// remove deletes the nodes at paths from their parents.
func remove(doc any, paths []spec.NormalizedPath) (any, error) {
	// Deepest and last first: removing an element of a sequence moves the
	// elements after it, and a node inside a removed node does not need to
	// be found afterwards.
	slices.SortFunc(paths, func(a, b spec.NormalizedPath) int { return b.Compare(a) })
	for _, p := range paths {
		if len(p) == 0 {
			return nil, errors.New("cannot remove the document root ($)")
		}
		parentPath, last := p[:len(p)-1], p[len(p)-1]
		parent, ok := resolve(doc, parentPath)
		if !ok {
			continue
		}
		switch sel := last.(type) {
		case spec.Name:
			if m, ok := parent.(map[string]any); ok {
				delete(m, string(sel))
			}
		case spec.Index:
			if s, ok := parent.([]any); ok && int(sel) < len(s) {
				doc = set(doc, parentPath, slices.Delete(s, int(sel), int(sel)+1))
			}
		}
	}
	return doc, nil
}

// update merges value into the nodes at paths.
func update(doc any, paths []spec.NormalizedPath, value any) (any, error) {
	for _, p := range paths {
		current, ok := resolve(doc, p)
		if !ok {
			continue
		}
		doc = set(doc, p, mergeTarget(current, clone(value)))
	}
	return doc, nil
}

// mergeTarget merges update into a node an action selected.
func mergeTarget(target, update any) any {
	if s, ok := target.([]any); ok {
		if us, ok := update.([]any); ok {
			return append(s, us...)
		}
		return append(s, update)
	}
	return merge(target, update)
}

// merge merges update into target: mappings merge recursively, sequences
// concatenate, and anything else is replaced by update.
func merge(target, update any) any {
	switch t := target.(type) {
	case map[string]any:
		u, ok := update.(map[string]any)
		if !ok {
			return update
		}
		for k, uv := range u {
			if tv, exists := t[k]; exists {
				t[k] = merge(tv, uv)
			} else {
				t[k] = uv
			}
		}
		return t
	case []any:
		if u, ok := update.([]any); ok {
			return append(t, u...)
		}
	}
	return update
}

// resolve returns the node at path.
func resolve(doc any, path spec.NormalizedPath) (any, bool) {
	node := doc
	for _, sel := range path {
		switch s := sel.(type) {
		case spec.Name:
			m, ok := node.(map[string]any)
			if !ok {
				return nil, false
			}
			if node, ok = m[string(s)]; !ok {
				return nil, false
			}
		case spec.Index:
			l, ok := node.([]any)
			if !ok || int(s) < 0 || int(s) >= len(l) {
				return nil, false
			}
			node = l[s]
		default:
			return nil, false
		}
	}
	return node, true
}

// set stores value at path, which must exist, and returns the document.
func set(doc any, path spec.NormalizedPath, value any) any {
	if len(path) == 0 {
		return value
	}
	parent, ok := resolve(doc, path[:len(path)-1])
	if !ok {
		return doc
	}
	switch s := path[len(path)-1].(type) {
	case spec.Name:
		if m, ok := parent.(map[string]any); ok {
			m[string(s)] = value
		}
	case spec.Index:
		if l, ok := parent.([]any); ok && int(s) < len(l) {
			l[s] = value
		}
	}
	return doc
}

// clone deep-copies a decoded JSON value, so that each target gets its own
// copy of an update.
func clone(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, e := range t {
			m[k] = clone(e)
		}
		return m
	case []any:
		l := make([]any, len(t))
		for i, e := range t {
			l[i] = clone(e)
		}
		return l
	default:
		return v
	}
}
