package externalref

import (
	"testing"

	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/externalref/gen/api"
	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/externalref/gen/common"
)

// From issue-2113: verify that a $ref to an external components/responses
// object correctly qualifies the schema type with the external package import.
func TestExternalRefInResponse(t *testing.T) {
	// This will fail to compile if the generated code uses ProblemDetails
	// instead of common.ProblemDetails (via the externalRef alias) in the
	// default response type.
	_ = api.ListThingsdefaultJSONResponse{
		Body:       common.ProblemDetails{Title: "err", Status: 500},
		StatusCode: 500,
	}
}
