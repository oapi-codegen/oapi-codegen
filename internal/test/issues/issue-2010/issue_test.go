package issue_2010_test

import (
	"testing"

	base "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-2010/gen/spec_base"
	other "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-2010/gen/spec_other"
)

// Cross-package cast that broke in 2.1.0+ when both specs generate
// strict-server. Compiling this file is the regression check: if the embedded
// field names diverge between the local and external strict envelopes, the
// conversion below fails to compile.
var _ = func(v base.GetExample400JSONResponse) other.GetOtherExample400JSONResponse {
	return other.GetOtherExample400JSONResponse(v)
}

func TestIssue2010ResponseCastAcrossPackages(t *testing.T) {
	var a base.GetExampleResponseObject = base.GetExample400JSONResponse{}
	switch v := a.(type) {
	case base.GetExample400JSONResponse:
		_ = other.GetOtherExample400JSONResponse(v)
	default:
		t.Fatalf("unexpected type %T", a)
	}
}
