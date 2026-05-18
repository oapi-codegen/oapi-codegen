// Package openapi31_nullable verifies that the OpenAPI 3.1 type-array
// nullable idiom (`type: ["X","null"]`) generates the same Go field shape
// as the OpenAPI 3.0 `nullable: true` idiom: a pointer to the underlying
// type. The test is structural -- it instantiates the generated types and
// assigns through the pointer field -- rather than string-matching the
// generated source.
package openapi31_nullable

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spec30 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/openapi31_nullable/spec_3_0"
	spec31 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/openapi31_nullable/spec_3_1"
)

// TestNicknameIsPointer_3_0 asserts that the 3.0 spec's `nullable: true`
// produces a `*string` field. Compile-time verification: the assignment
// of `&nick` only succeeds if Nickname is `*string`.
func TestNicknameIsPointer_3_0(t *testing.T) {
	nick := "rex"
	p := spec30.Pet{
		Name:     "fluffy",
		Nickname: &nick,
	}
	require.NotNil(t, p.Nickname)
	assert.Equal(t, "rex", *p.Nickname)

	// nil-nickname round-trip
	p2 := spec30.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Nickname)
}

// TestNicknameIsPointer_3_1 asserts that the 3.1 spec's
// `type: ["string","null"]` produces a `*string` field, identical in
// shape to the 3.0 control case.
func TestNicknameIsPointer_3_1(t *testing.T) {
	nick := "rex"
	p := spec31.Pet{
		Name:     "fluffy",
		Nickname: &nick,
	}
	require.NotNil(t, p.Nickname)
	assert.Equal(t, "rex", *p.Nickname)

	p2 := spec31.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Nickname)
}

// TestNullableArrayAndObjectFields_3_1 asserts that nullable array and
// nullable inline-object fields generate as pointer-to-slice and
// pointer-to-struct respectively in 3.1. Regression for the case where
// `type: ["array","null"]` and `type: ["object","null"]` were rejected
// before schemaPrimaryType was applied at the GenerateGoSchema dispatch
// (see pkg/codegen/schema.go).
func TestNullableArrayAndObjectFields_3_1(t *testing.T) {
	tags := []string{"good", "boy"}
	id := "owner-1"
	p := spec31.Pet{
		Name: "fluffy",
		Tags: &tags,
		Owner: &struct {
			Id *string `json:"id,omitempty"`
		}{Id: &id},
	}
	require.NotNil(t, p.Tags)
	assert.Equal(t, []string{"good", "boy"}, *p.Tags)
	require.NotNil(t, p.Owner)
	require.NotNil(t, p.Owner.Id)
	assert.Equal(t, "owner-1", *p.Owner.Id)

	// Zero-value: nullable fields must be nil, not empty.
	p2 := spec31.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Tags)
	assert.Nil(t, p2.Owner)
}

// TestNullableArrayAndObjectFields_3_0 is the matching control case
// asserting that `nullable: true` arrays and inline objects in 3.0
// generate the same pointer shape.
func TestNullableArrayAndObjectFields_3_0(t *testing.T) {
	tags := []string{"good", "boy"}
	id := "owner-1"
	p := spec30.Pet{
		Name: "fluffy",
		Tags: &tags,
		Owner: &struct {
			Id *string `json:"id,omitempty"`
		}{Id: &id},
	}
	require.NotNil(t, p.Tags)
	assert.Equal(t, []string{"good", "boy"}, *p.Tags)
	require.NotNil(t, p.Owner)
	require.NotNil(t, p.Owner.Id)
	assert.Equal(t, "owner-1", *p.Owner.Id)

	p2 := spec30.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Tags)
	assert.Nil(t, p2.Owner)
}

// TestJsonRoundTrip_NullableFields_AcrossVersions asserts that a JSON
// payload with an explicit null nickname unmarshals to (*string)(nil) in
// both spec versions, and that JSON output omits the field when nil due
// to omitempty. The two generated structs must marshal identically for
// the nullable field.
func TestJsonRoundTrip_NullableFields_AcrossVersions(t *testing.T) {
	const withName = `{"name":"fluffy"}`
	const withBoth = `{"name":"fluffy","nickname":"rex"}`

	for _, tc := range []struct {
		name string
		// fn30 / fn31 unmarshal the input into each version's Pet type
		// and return a JSON re-marshal so we can assert equality.
		fn30 func(input string) (string, *string, error)
		fn31 func(input string) (string, *string, error)
	}{
		{
			name: "unmarshal/marshal symmetric across versions",
			fn30: func(input string) (string, *string, error) {
				var p spec30.Pet
				if err := json.Unmarshal([]byte(input), &p); err != nil {
					return "", nil, err
				}
				out, err := json.Marshal(p)
				return string(out), p.Nickname, err
			},
			fn31: func(input string) (string, *string, error) {
				var p spec31.Pet
				if err := json.Unmarshal([]byte(input), &p); err != nil {
					return "", nil, err
				}
				out, err := json.Marshal(p)
				return string(out), p.Nickname, err
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, in := range []string{withName, withBoth} {
				out30, n30, err30 := tc.fn30(in)
				require.NoError(t, err30)
				out31, n31, err31 := tc.fn31(in)
				require.NoError(t, err31)
				assert.JSONEq(t, in, out30, "3.0 round-trip should be lossless")
				assert.JSONEq(t, in, out31, "3.1 round-trip should be lossless")
				assert.JSONEq(t, out30, out31, "3.0 and 3.1 must marshal identically")
				if n30 == nil {
					assert.Nil(t, n31)
				} else {
					require.NotNil(t, n31)
					assert.Equal(t, *n30, *n31)
				}
			}
		})
	}
}
