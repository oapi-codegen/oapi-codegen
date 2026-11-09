// Package responsescomponentinlineunion verifies that a components/responses
// entry whose inline schema is a oneOf/anyOf union generates a compilable
// strict-server envelope when an operation references it via $ref. The
// strict <Name>JSONResponse envelope must point at the union type declared for
// the component (BadRequest), not at a synthetic <Name>JSONResponseBody type
// that nothing declares, and the Visit method must be able to reach the
// union's raw JSON through the embedded envelope.
//
// From issue-2539.
package responsescomponentinlineunion

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
