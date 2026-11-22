package responsescontent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// When a components/responses entry exposes the same schema under more than one
// content type (here application/json + application/xml), the client response
// wrapper grows one field per content type (JSON500, XML500). The regression in
// issue #2389 typed those fields as undefined per-content-type names
// (*ErrorResponseApplicationJSON / *ErrorResponseApplicationXML), so the
// generated package failed to compile. Both fields must instead point at the
// single declared component type, ErrorResponse.
//
// This test exists primarily to prove the generated package compiles; the
// assignments below would not type-check if either field had a different (or
// undefined) type.
func TestResponseComponentMultipleContentTypesShareDeclaredType(t *testing.T) {
	body := &ErrorResponse{}

	resp := GetThingResponse{
		JSON500: body,
		XML500:  body,
	}

	assert.Same(t, resp.JSON500, resp.XML500)
}

// A component response only declares a Go type for its JSON content, so an
// XML-only component response has no declared base type. The wrapper field must
// point at the XML content's own schema type (XmlError); pointing at the
// component base type would reference an undeclared name and fail to compile.
func TestResponseComponentXMLOnlyUsesContentSchemaType(t *testing.T) {
	resp := GetXMLOnlyResponse{
		XML500: &XmlError{},
	}

	assert.NotNil(t, resp.XML500)
}

// When the JSON and XML content of a component response resolve to different
// schemas, the two wrapper fields must keep distinct types: JSON uses the
// declared component type (MixedError) and XML keeps its own schema type
// (XmlError). The regression shared the JSON type for both, which would
// silently decode XML into the JSON-shaped type. These statements would not
// type-check if the fields shared a type.
func TestResponseComponentMixedDifferingSchemasKeepDistinctTypes(t *testing.T) {
	resp := GetMixedResponse{
		JSON500: &MixedError{},
		XML500:  &XmlError{},
	}

	assert.NotNil(t, resp.JSON500)
	assert.NotNil(t, resp.XML500)
}
