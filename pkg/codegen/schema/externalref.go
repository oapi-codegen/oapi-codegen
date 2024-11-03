package schema

import (
	"fmt"
	"strings"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/singleton"
)

// EnsureExternalRefsInRequestBodyDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func EnsureExternalRefsInRequestBodyDefinitions(defs *[]RequestBodyDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, rbd := range *defs {
		EnsureExternalRefsInSchema(&rbd.Schema, ref)

		// make sure we then update it in-place
		(*defs)[i] = rbd
	}
}

// EnsureExternalRefsInResponseDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func EnsureExternalRefsInResponseDefinitions(defs *[]ResponseDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, rd := range *defs {
		for j, rcd := range rd.Contents {
			EnsureExternalRefsInSchema(&rcd.Schema, ref)

			// make sure we then update it in-place
			rd.Contents[j] = rcd
		}

		// make sure we then update it in-place
		(*defs)[i] = rd
	}
}

// EnsureExternalRefsInParameterDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func EnsureExternalRefsInParameterDefinitions(defs *[]ParameterDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, pd := range *defs {
		EnsureExternalRefsInSchema(&pd.Schema, ref)

		// make sure we then update it in-place
		(*defs)[i] = pd
	}
}

// EnsureExternalRefsInSchema ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
//
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
//
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func EnsureExternalRefsInSchema(schema *Schema, ref string) {
	if ref == "" {
		return
	}

	// if this is already defined as the start of a struct, we shouldn't inject **??**
	if strings.HasPrefix(schema.GoType, "struct {") {
		return
	}

	parts := strings.SplitN(ref, "#", 2)
	if pack, ok := singleton.GlobalState.ImportMapping[parts[0]]; ok {
		schema.RefType = fmt.Sprintf("%s.%s", pack.Name, schema.GoType)
	}
}
