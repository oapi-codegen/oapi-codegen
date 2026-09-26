package codegen

import (
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

// genContext carries the state that has to survive the mutual recursion
// between generateGoSchema and the allOf code of every schema-merging-behavior
// version (merge_schemas.go, merge_schemas_v2.go, merge_schemas_v3.go).
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
	// generated for that allOf's merged result. generateAllOfV2 and
	// generateAllOfV3 register an entry before generating the merged body and
	// remove it afterwards, so a member that refers back into a body an
	// enclosing frame is still generating resolves to that type's name (see
	// inProgressAllOf) instead of being inlined again, which is what used to
	// recurse forever (issue #2542).
	//
	// The key is the schema node that owns the allOf, not the allOf slice:
	// both versions rebuild that slice on every call, to merge the schema's
	// own keywords as a member, so only the owner is stable across re-entry.
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
	// cannot name itself, so finishAllOf reports the shape instead of
	// guessing. No spec is known to reach that: a $ref returns before the
	// allOf block, and the only other way back to a component's own allOf
	// node is a copy of it that the merges flatten rather than hand to
	// generateGoSchema. v3 also reads it to keep a component from being
	// defined as an alias of itself (see annotatedRefMember).
	rootPosition bool

	// memberLabels names the allOf members the generator makes up rather
	// than reads from the spec, for error messages that name the members of
	// a composition: the parent's own keywords, which generateAllOfV3 merges
	// as a member, are the schema itself (""). Only schema-merging-behavior
	// v3 uses it.
	memberLabels map[*openapi3.SchemaRef]string

	// madeUp records the allOfs v3's merge makes for positions several
	// members declare, which the spec doesn't spell out.
	madeUp map[*openapi3.Schema]bool

	// unionComponents holds, for a schema v3's merge made from an allOf with
	// several oneOfs or anyOfs, those unions (see unionComponent).
	unionComponents map[*openapi3.Schema][]unionComponent

	// subschemas holds the allOfs v3's merge makes of the schemas several
	// members declare for one position, by those schemas, so the same ones
	// always make the same allOf (see allOfMerge.subschema).
	subschemas map[string]*openapi3.SchemaRef

	// run holds settings of the generation run that the walk reads (see
	// runSettings).
	run runSettings

	// variantKeyCache holds the JSON keys of the union variants v3 has read
	// them for, by schema (see genContext.variantKeys).
	variantKeyCache map[*openapi3.Schema][]string
}

// mergeFrame describes an allOf merge that an enclosing frame is part-way
// through generating.
type mergeFrame struct {
	// typeName is the Go type the merged result will be defined as, if the
	// composition turns out to be recursive.
	typeName string
	// consulted records whether anything below actually referred back to
	// this merge. When nothing did, the composition is not recursive and
	// finishAllOf returns the inline anonymous struct it always has, so
	// non-recursive output stays byte-identical.
	consulted bool
}

// runSettings are the settings of the generation run that the schema walk
// reads as it goes. newGenContext reads them from the configuration once, so
// the walk doesn't consult globalState for each schema.
type runSettings struct {
	// merging is the schema-merging-behavior in effect. Generate has already
	// rejected options that don't select one; v2 stands in for them here.
	merging string
	// lenientUnionAccessors is output-options.lenient-union-accessors.
	lenientUnionAccessors bool
	// typesForAnonymousSchemas is
	// output-options.generate-types-for-anonymous-schemas.
	typesForAnonymousSchemas bool
}

// currentRunSettings reads the runSettings of the run's configuration.
func currentRunSettings() runSettings {
	merging, err := globalState.options.Compatibility.schemaMergingVersion()
	if err != nil {
		merging = SchemaMergingV2
	}
	return runSettings{
		merging:                  merging,
		lenientUnionAccessors:    globalState.options.OutputOptions.LenientUnionAccessors,
		typesForAnonymousSchemas: globalState.options.OutputOptions.GenerateTypesForAnonymousSchemas,
	}
}

// newGenContext returns a context rooted at a top-level schema position.
func newGenContext(nameHint []string) genContext {
	return genContext{
		run:             currentRunSettings(),
		inProgress:      make(map[*openapi3.Schema]*mergeFrame),
		nameHint:        slices.Clone(nameHint),
		memberLabels:    make(map[*openapi3.SchemaRef]string),
		subschemas:      make(map[string]*openapi3.SchemaRef),
		madeUp:          make(map[*openapi3.Schema]bool),
		unionComponents: make(map[*openapi3.Schema][]unionComponent),
		variantKeyCache: make(map[*openapi3.Schema][]string),
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
// finishAllOf re-checks it against whatever actually got defined.
//
// PathToTypeName rewrites the slice it is handed, so it gets a copy.
func (ctx genContext) typeName(path []string) string {
	if ctx.run.typesForAnonymousSchemas && len(path) > 1 {
		return PathToTypeName(slices.Clone(path))
	}
	return PathToTypeName(slices.Clone(ctx.nameHint))
}
