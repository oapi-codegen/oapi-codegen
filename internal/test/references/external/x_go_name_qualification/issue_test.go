package xgonamequalification

import (
	"testing"

	spec_base "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/external/x_go_name_qualification/gen/spec_base"
	spec_ext "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/external/x_go_name_qualification/gen/spec_ext"
)

// Regression test for https://github.com/oapi-codegen/oapi-codegen/issues/2422.
//
// The base spec references an external response directly. That response is
// renamed via x-go-name in the external package (Outcome -> OutcomeResult), so
// the generated client wrapper field must be typed as the imported model name.
//
// Assigning an external-typed value into the wrapper field only compiles when
// the field is exactly *spec_ext.OutcomeResult: two distinct named types are not
// assignable to each other, so this pins the field's type to the external
// package's model.
func TestDirectExternalResponseHonoursXGoName(t *testing.T) {
	outcome := spec_ext.OutcomeResult("ok")
	resp := spec_base.GetOutcomeResponse{
		JSON200: &outcome,
	}
	_ = resp
}
