package codegen

import (
	"fmt"
	"strings"
)

// ensureExternalRefsInRequestBodyDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func ensureExternalRefsInRequestBodyDefinitions(defs *[]RequestBodyDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, rbd := range *defs {
		ensureExternalRefsInSchema(&rbd.Schema, ref)

		// make sure we then update it in-place
		(*defs)[i] = rbd
	}
}

// ensureExternalRefsInResponseDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func ensureExternalRefsInResponseDefinitions(defs *[]ResponseDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, rd := range *defs {
		for j, rcd := range rd.Contents {
			ensureExternalRefsInSchema(&rcd.Schema, ref)

			// make sure we then update it in-place
			rd.Contents[j] = rcd
		}

		// make sure we then update it in-place
		(*defs)[i] = rd
	}
}

// ensureExternalRefsInParameterDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func ensureExternalRefsInParameterDefinitions(defs *[]ParameterDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, pd := range *defs {
		ensureExternalRefsInSchema(&pd.Schema, ref)

		// make sure we then update it in-place
		(*defs)[i] = pd
	}
}

// ensureExternalRefsInSchema ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
//
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
//
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func ensureExternalRefsInSchema(schema *Schema, ref string) {
	if ref == "" {
		return
	}

	// if this is already defined as the start of a struct, we shouldn't inject **??**
	if strings.HasPrefix(schema.GoType, "struct {") {
		return
	}

	parts := strings.SplitN(ref, "#", 2)
	if pack, ok := globalState.importMapping[parts[0]]; ok {
		schema.RefType = fmt.Sprintf("%s.%s", pack.Name, schema.GoType)
	}
}
