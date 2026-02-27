package issue2113

import (
	"testing"

	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-2113/gen/api"
	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-2113/gen/common"
)

// TestExternalRefInResponse verifies that a $ref to an external
// components/responses object correctly qualifies the schema type
// with the external package import. See
// https://github.com/oapi-codegen/oapi-codegen/issues/2113
func TestExternalRefInResponse(t *testing.T) {
	// This will fail to compile if the generated code uses
	// ProblemDetails instead of common.ProblemDetails (via the
	// externalRef alias) in the default response type.
	_ = api.ListThingsdefaultJSONResponse{
		Body:       common.ProblemDetails{Title: "err", Status: 500},
		StatusCode: 500,
	}
}
