// Package openapi31_content_keywords exercises the codegen mapping for
// OpenAPI 3.1's `contentEncoding` and `contentMediaType` keywords. The
// assertions are structural -- assigning typed zero values into the
// generated fields only compiles if the field types are what we
// expect, so a regression that drops the synthesis would surface as a
// build failure here before any test run.
package openapi31_content_keywords

import (
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"

	spec "github.com/oapi-codegen/oapi-codegen/v2/internal/test/openapi31_content_keywords/spec"
)

func TestContentMediaTypeMapsToFile(t *testing.T) {
	// `contentMediaType: application/octet-stream` -- raw binary --
	// must produce openapi_types.File, matching 3.0 `format: binary`.
	var f openapi_types.File
	fields := spec.FileUploadFields{
		OrderId: 1,
		RawFile: f,
	}
	assert.Equal(t, 1, fields.OrderId)
}

func TestContentMediaTypeImageMapsToFile(t *testing.T) {
	// A non-octet-stream contentMediaType (image/png) still routes to
	// openapi_types.File: the keyword signals "this string is binary
	// content," regardless of the specific media type.
	var f openapi_types.File
	fields := spec.FileUploadFields{
		ImageFile: &f,
	}
	assert.NotNil(t, fields.ImageFile)
}

func TestContentEncodingMapsToByteSlice(t *testing.T) {
	// `contentEncoding: base64` (and other RFC4648 encodings) must
	// produce []byte, matching 3.0 `format: byte` behavior.
	fields := spec.FileUploadFields{
		Base64Field:    []byte("hello"),
		Base64UrlField: []byte("world"),
	}
	assert.Equal(t, "hello", string(fields.Base64Field))
	assert.Equal(t, "world", string(fields.Base64UrlField))
}

func TestContentMediaTypeWinsOverContentEncoding(t *testing.T) {
	// When both keywords are set, the file mapping wins -- the field
	// is a binary blob of a specific media type, base64-encoded for
	// JSON transport. Matches the 3.0 model where `format: binary`
	// dominated.
	var f openapi_types.File
	fields := spec.FileUploadFields{
		BothKeywords: &f,
	}
	assert.NotNil(t, fields.BothKeywords)
}

func TestExplicitFormatOverridesContentMediaType(t *testing.T) {
	// An explicit `format` on the schema must continue to win over
	// the contentMediaType-derived synthesis. Here `format: date`
	// keeps its openapi_types.Date Go type even though
	// contentMediaType: application/octet-stream would otherwise have
	// routed to openapi_types.File.
	var d openapi_types.Date
	fields := spec.FileUploadFields{
		ExplicitFormatWins: &d,
	}
	assert.NotNil(t, fields.ExplicitFormatWins)
}
