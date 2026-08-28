package responsecast_test

import (
	"mime/multipart"
	"strings"
	"testing"

	"github.com/google/uuid"

	base "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/response_cast/gen/spec_base"
	other "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/response_cast/gen/spec_other"
	otherfiber "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/response_cast/gen/spec_other_fiber"
	otheriris "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/response_cast/gen/spec_other_iris"
	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/response_cast/support"
)

// Cross-package cast that broke in 2.1.0+ when both specs generate
// strict-server. Compiling this file is the regression check: if the embedded
// field names diverge between the local and external strict envelopes, the
// conversion below fails to compile.
var _ = func(v base.GetExample400JSONResponse) other.GetOtherExample400JSONResponse {
	return other.GetOtherExample400JSONResponse(v)
}

// Reusable response envelopes must remain convertible across strict-server
// template families, including bodies which require a Body field.
var (
	_ = func(v base.GetExample400JSONResponse) otherfiber.GetOtherExample400JSONResponse {
		return otherfiber.GetOtherExample400JSONResponse(v)
	}
	_ = func(v base.GetExample400JSONResponse) otheriris.GetOtherExample400JSONResponse {
		return otheriris.GetOtherExample400JSONResponse(v)
	}
	_ = func(v base.GetExample401JSONResponse) other.GetOtherExample401JSONResponse {
		return other.GetOtherExample401JSONResponse(v)
	}
	_ = func(v base.GetExample401JSONResponse) otherfiber.GetOtherExample401JSONResponse {
		return otherfiber.GetOtherExample401JSONResponse(v)
	}
	_ = func(v base.GetExample401JSONResponse) otheriris.GetOtherExample401JSONResponse {
		return otheriris.GetOtherExample401JSONResponse(v)
	}
	_ = func(v base.GetExample402JSONResponse) other.GetOtherExample402JSONResponse {
		return other.GetOtherExample402JSONResponse(v)
	}
	_ = func(v base.GetExample402JSONResponse) otherfiber.GetOtherExample402JSONResponse {
		return otherfiber.GetOtherExample402JSONResponse(v)
	}
	_ = func(v base.GetExample402JSONResponse) otheriris.GetOtherExample402JSONResponse {
		return otheriris.GetOtherExample402JSONResponse(v)
	}
	_ = func(v base.GetExample403JSONResponse) other.GetOtherExample403JSONResponse {
		return other.GetOtherExample403JSONResponse(v)
	}
	_ = func(v base.GetExample403JSONResponse) otherfiber.GetOtherExample403JSONResponse {
		return otherfiber.GetOtherExample403JSONResponse(v)
	}
	_ = func(v base.GetExample403JSONResponse) otheriris.GetOtherExample403JSONResponse {
		return otheriris.GetOtherExample403JSONResponse(v)
	}
)

var (
	_ = base.N401JSONResponse{Body: map[string]any{"source": "base"}}
	_ = base.N402JSONResponse{Body: new(string)}
	_ = base.N403JSONResponse{
		Body:    map[string]any{"source": "base"},
		Headers: base.N403ResponseHeaders{XRequestID: "request-id"},
	}
	_ = func(v base.N400) base.GetExample404JSONResponse {
		return base.GetExample404JSONResponse(v)
	}
)

// Resolved external aliases are classified by their underlying schema.
var (
	_ = other.GetOtherExample404JSONResponse{Body: map[string]any{"source": "other"}}
	_ = otherfiber.GetOtherExample404JSONResponse{Body: map[string]any{"source": "fiber"}}
	_ = otheriris.GetOtherExample404JSONResponse{Body: map[string]any{"source": "iris"}}
	_ = other.GetOtherExample405JSONResponse{Body: nil}
	_ = otherfiber.GetOtherExample405JSONResponse{Body: nil}
	_ = otheriris.GetOtherExample405JSONResponse{Body: nil}
	_ = other.GetOtherExample406JSONResponse{Body: new(string)}
	_ = otherfiber.GetOtherExample406JSONResponse{Body: new(string)}
	_ = otheriris.GetOtherExample406JSONResponse{Body: new(string)}
)

// External path-item multipart responses keep their callback receiver in each
// strict-server template family.
var (
	_ other.ExternalMultipartResponse200MultipartResponse      = func(*multipart.Writer) error { return nil }
	_ otherfiber.ExternalMultipartResponse200MultipartResponse = func(*multipart.Writer) error { return nil }
	_ otheriris.ExternalMultipartResponse200MultipartResponse  = func(*multipart.Writer) error { return nil }
)

// Opaque external x-go-type selectors preserve their direct conversion API.
var (
	_ = func(v uuid.UUID) other.GetOtherExample407JSONResponse {
		return other.GetOtherExample407JSONResponse(v)
	}
	_ = func(v uuid.UUID) otherfiber.GetOtherExample407JSONResponse {
		return otherfiber.GetOtherExample407JSONResponse(v)
	}
	_ = func(v uuid.UUID) otheriris.GetOtherExample407JSONResponse {
		return otheriris.GetOtherExample407JSONResponse(v)
	}
	_ = func(v support.GenericResponse[string]) other.GetOtherExample408JSONResponse {
		return other.GetOtherExample408JSONResponse(v)
	}
	_ = func(v support.GenericResponse[string]) otherfiber.GetOtherExample408JSONResponse {
		return otherfiber.GetOtherExample408JSONResponse(v)
	}
	_ = func(v support.GenericResponse[string]) otheriris.GetOtherExample408JSONResponse {
		return otheriris.GetOtherExample408JSONResponse(v)
	}
)

// Unsupported response media types preserve their io.Reader envelope.
var (
	_ = other.GetOtherExample409ApplicationoctetStreamResponse{Body: strings.NewReader("other"), ContentLength: 5}
	_ = otherfiber.GetOtherExample409ApplicationoctetStreamResponse{Body: strings.NewReader("fiber"), ContentLength: 5}
	_ = otheriris.GetOtherExample409ApplicationoctetStreamResponse{Body: strings.NewReader("iris"), ContentLength: 4}
)

func TestResponseCastAcrossPackages(t *testing.T) {
	var a base.GetExampleResponseObject = base.GetExample400JSONResponse{}
	switch v := a.(type) {
	case base.GetExample400JSONResponse:
		_ = other.GetOtherExample400JSONResponse(v)
	default:
		t.Fatalf("unexpected type %T", a)
	}
}
