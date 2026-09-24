package codegen

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// MergeSchemas merges all the fields in the schemas supplied into one giant
// schema. The idea is that we merge all fields together into one schema.
//
// It starts a fresh generation context; within the package, the allOf code
// is handed its schemaMerger by generateAllOf instead.
func MergeSchemas(allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	merge := schemaMergerFor(schemaMergingInEffect())
	return merge(newGenContext(path), allOf, path)
}

// schemaMergingInEffect returns the schema-merging-behavior of the current
// run. Generate has already rejected options that don't select one; v2 stands
// in for them here.
func schemaMergingInEffect() string {
	version, err := globalState.options.Compatibility.schemaMergingVersion()
	if err != nil {
		return SchemaMergingV2
	}
	return version
}

// schemaMerger merges the members of an allOf into one Schema.
type schemaMerger func(ctx genContext, allOf []*openapi3.SchemaRef, path []string) (Schema, error)

// schemaMergerFor returns the allOf merge of a schema-merging-behavior
// version.
func schemaMergerFor(version string) schemaMerger {
	switch version {
	case SchemaMergingV1:
		return func(_ genContext, allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
			return mergeSchemasV1(allOf, path)
		}
	case SchemaMergingV2:
		return mergeSchemasV2
	case schemaMergingV3:
		return mergeSchemasV3
	default:
		// schemaMergingVersion only returns the versions above; a new one
		// has to be added here too.
		panic(fmt.Sprintf("no allOf merge for schema-merging-behavior %q", version))
	}
}

// generateAllOf lowers a schema with allOf for generateGoSchema, the way the
// schema-merging-behavior in effect does. v1 and v2 share v2's code, which is
// handed the version's own merge of the members.
func generateAllOf(ctx genContext, schema *openapi3.Schema, path []string, extensions map[string]any, skipOptionalPointer bool) (Schema, error) {
	version := schemaMergingInEffect()
	merge := schemaMergerFor(version)
	if version == schemaMergingV3 {
		return generateAllOfV3(ctx, schema, path, extensions, skipOptionalPointer, merge)
	}
	return generateAllOfV2(ctx, schema, path, extensions, skipOptionalPointer, merge)
}

// nonNullTypes returns a type array's entries other than "null".
func nonNullTypes(t *openapi3.Types) []string {
	var out []string
	for _, typ := range t.Slice() {
		if typ != openapi3.TypeNull {
			out = append(out, typ)
		}
	}
	return out
}
