package codegen

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// genContext carries the state that has to survive the mutual recursion
// between generateGoSchema and the allOf code in merge_schemas_v2.go.
//
// The rest of V2 keeps generator state in globalState, but this state is
// positional rather than per-run: it describes where inside one schema's
// expansion we currently are. Threading it keeps generateGoSchema reentrant.
//
// A genContext is passed by value. The map inside it is shared by every copy
// — that is the point of inProgress — while the scalar fields are per-frame,
// so descending can deepen them without the caller having to restore them.
type genContext struct {
	// inProgress maps a schema node that owns an allOf to the type being
	// generated for that allOf's merged result. generateAllOfV2 registers an
	// entry before generating the merged body and removes it afterwards, so
	// a member that refers back into a body an enclosing frame is still
	// generating resolves to that type's name instead of being inlined
	// again, which is what used to recurse forever (issue #2542).
	//
	// The key is the schema node that owns the allOf, not the allOf slice:
	// the sibling-injection path in generateAllOfV2 rebuilds that slice on
	// every call, so only the owner is stable across re-entry.
	inProgress map[*openapi3.Schema]*mergeFrame

	// nameHint is the path a type generated at this position would be named
	// from — the argument the calling frame would hand to PathToTypeName.
	// It is deliberately not `path`: items and additionalProperties reuse
	// their parent's path but name their types with an extra element, and
	// changing that would rename nested types across every spec.
	nameHint []string

	// rootPosition marks the top of a components/schemas entry, where
	// GenerateTypesForSchemas defines the type under a name renameSchema
	// chooses rather than one derived from the path. A composition there
	// cannot name itself, so generateAllOfV2 reports the shape instead of
	// guessing. No spec is known to reach it — a $ref returns before the
	// allOf block, and the only other way back to a component's own allOf
	// node is valueWithPropagatedRefV2's copy, which mergeAllOfV2 flattens
	// rather than handing to generateGoSchema — so this is a diagnostic for
	// a case believed impossible, not a supported path.
	rootPosition bool
}

// mergeFrame describes an allOf merge that an enclosing frame is part-way
// through generating.
type mergeFrame struct {
	// typeName is the Go type the merged result will be defined as, if the
	// composition turns out to be recursive.
	typeName string
	// consulted records whether anything below actually referred back to
	// this merge. When nothing did, the composition is not recursive and
	// generateAllOfV2 returns the inline anonymous struct it always has, so
	// non-recursive output stays byte-identical.
	consulted bool
}

// newGenContext returns a context rooted at a top-level schema position.
func newGenContext(nameHint []string) genContext {
	return genContext{
		inProgress: make(map[*openapi3.Schema]*mergeFrame),
		nameHint:   slices.Clone(nameHint),
	}
}

// newRootGenContext returns a context rooted at a components/schemas entry.
func newRootGenContext(nameHint []string) genContext {
	ctx := newGenContext(nameHint)
	ctx.rootPosition = true
	return ctx
}

// at returns a copy of ctx positioned at nameHint. Descending always leaves
// the caller-names-it position behind.
func (ctx genContext) at(nameHint []string) genContext {
	ctx.nameHint = slices.Clone(nameHint)
	ctx.rootPosition = false
	return ctx
}

// typeName is the Go type a type generated at this position is called.
//
// The name has to be settled before the body is generated, because a
// reference handed to a recursive member needs it, so this predicts what will
// end up defining the type: under generate-types-for-anonymous-schemas
// generateGoSchema hoists an anonymous object under its own path and that
// name wins, otherwise the property, items and additionalProperties naming
// blocks name it after this position. The anonymous-schema rule also depends
// on the generated result, so the prediction is deliberately loose and
// generateAllOfV2 re-checks it against whatever actually got defined.
//
// PathToTypeName rewrites the slice it is handed, so it gets a copy.
func (ctx genContext) typeName(path []string) string {
	if globalState.options.OutputOptions.GenerateTypesForAnonymousSchemas && len(path) > 1 {
		return PathToTypeName(slices.Clone(path))
	}
	return PathToTypeName(slices.Clone(ctx.nameHint))
}
