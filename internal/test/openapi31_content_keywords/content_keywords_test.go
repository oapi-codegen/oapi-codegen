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
	"github.com/stretchr/testify/require"

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

func TestContentEncodingBase64MapsToByteSlice(t *testing.T) {
	// `contentEncoding: base64` (standard padded base64) is the one
	// RFC4648 variant Go's encoding/json handles natively for []byte,
	// so it maps to []byte -- matching 3.0 `format: byte`.
	fields := spec.FileUploadFields{
		Base64Field: []byte("hello"),
	}
	assert.Equal(t, "hello", string(fields.Base64Field))
}

func TestContentEncodingBase64UrlStaysString(t *testing.T) {
	// `contentEncoding: base64url` (URL-safe base64) is intentionally
	// NOT mapped to []byte: Go's JSON codec for []byte uses standard
	// base64 unconditionally, which would silently corrupt URL-safe
	// characters (`-`/`_`) on unmarshal and re-emit them as standard
	// base64 on marshal. The field stays as `string` so the declared
	// wire encoding is preserved end-to-end and the user can apply
	// the correct codec. Compile-time check: `*string` assignment.
	v := "abc-def_ghi"
	fields := spec.FileUploadFields{
		Base64UrlField: &v,
	}
	require.NotNil(t, fields.Base64UrlField)
	assert.Equal(t, "abc-def_ghi", *fields.Base64UrlField)
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
