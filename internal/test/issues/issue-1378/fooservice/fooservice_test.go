package fooservice

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	bionicle "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-1378/bionicle"
)

// TestExternalRefUnionResponseSerialization locks in the fix for the
// case where a strict-server response envelope wraps an external union
// type. Because the strict envelope is a defined type (`type X
// externalRef0.Y`), methods on Y don't transfer; without an explicit
// MarshalJSON delegator the encode falls back to struct-field
// serialization on the unexported `union` field and produces `{}`.
func TestExternalRefUnionResponseSerialization(t *testing.T) {
	var body bionicle.GetBionicleName400JSONResponseBody
	require.NoError(t, body.FromBionicle(bionicle.Bionicle{Name: "tahu"}))

	resp := GetBionicleName400JSONResponse(body)

	got, err := json.Marshal(resp)
	require.NoError(t, err)
	require.JSONEq(t, `{"name":"tahu"}`, string(got))
}
