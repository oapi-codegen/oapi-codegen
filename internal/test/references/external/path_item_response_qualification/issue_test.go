package pathitemresponsequalification

import (
	"testing"

	spec_base "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/external/path_item_response_qualification/gen/spec_base"
	spec_ext "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/external/path_item_response_qualification/gen/spec_ext"
)

// Regression test for https://github.com/oapi-codegen/oapi-codegen/issues/2308.
//
// The base spec references an external path item (spec-ext.yaml) whose 200
// response is, in turn, a ref relative to that external file. The embedded
// response struct in the base package must therefore be qualified with the
// external package, i.e. externalRef0.VersionGetResponseJSONResponse.
//
// Before the fix the generator emitted an unqualified VersionGetResponseJSONResponse,
// which does not exist in the base package, so spec_base failed to compile. This
// test only needs to build: the assignment below pins the embedded field's type
// to the external package's type.
func TestExternalPathItemResponseIsQualified(t *testing.T) {
	// Setting the embedded field to an external-typed value only compiles when
	// the embed is exactly externalRef0.VersionGetResponseJSONResponse: two
	// distinct named types are not assignable to each other, so this pins the
	// field's type to the external package's.
	resp := spec_base.GetV1Version200JSONResponse{
		VersionGetResponseJSONResponse: spec_ext.VersionGetResponseJSONResponse("v1.2.3"),
	}
	_ = resp
}
