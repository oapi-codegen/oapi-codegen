package aggregateshoisting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- explicit hoisting (generate-types-for-anonymous-schemas) ----
// Source: anonymous_inner_hoisting/global

// TestHoistedTypesExist verifies that with output-options.generate-types-for-
// anonymous-schemas enabled, the inline schemas in spec_explicit.yaml become named
// Go types we can reference directly. The spec is the canonical issue #1139
// shape: a response body using `allOf` to merge a $ref with sibling
// `properties:` containing an inline `data` object.
func TestHoistedTypesExist(t *testing.T) {
	// Both the response root and the nested inline `data` schema should be
	// emitted as named types — assigning a typed zero value would not
	// compile if either were still anonymous structs.
	var responseBody GetRolesId200JSONResponseBody
	var dataField GetRolesId200JSONResponseBody_Data

	// Field-level type identity: GetRolesId200JSONResponseBody.Data must be
	// of the hoisted GetRolesId200JSONResponseBody_Data type. This
	// assignment fails to compile if Data is still an anonymous struct.
	responseBody.Data = dataField
	_ = responseBody
}

func TestHoistedTypesRoundTrip(t *testing.T) {
	body := GetRolesId200JSONResponseBody{
		Data: GetRolesId200JSONResponseBody_Data{
			Role: Role{Id: 7, Name: "admin"},
		},
		Ok: true,
	}

	encoded, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded GetRolesId200JSONResponseBody
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, body, decoded)
}

// TestArrayOfInlineObjectsHoisted verifies that a top-level array schema
// whose items are an inline object emits a named element type under the
// flag, instead of `type Roles = []struct{...}`. The element type's name
// is Roles_Item (path + "Item" suffix, matching the existing array-item
// hoist convention for unions and additionalProperties).
func TestArrayOfInlineObjectsHoisted(t *testing.T) {
	// Compile-time assertion: Roles is a slice whose element type is the
	// named Roles_Item type, not an anonymous struct.
	roles := Roles{
		Roles_Item{Id: 1, Name: "admin"},
		Roles_Item{Id: 2, Name: "user"},
	}
	require.Len(t, roles, 2)
	assert.Equal(t, "admin", roles[0].Name)
}

// ---- implicit/default hoisting (no flag; OpenAPI 3.1) ----
// Source: anonymous_inner_hoisting/implicit

func ptr[T any](v T) *T { return &v }

func TestResponseRootOneOf_RoundTripCat(t *testing.T) {
	cat := ImplicitCat{Kind: Cat, Name: ptr("whiskers")}

	var u GetResponseRootOneOf200JSONResponseBody
	require.NoError(t, u.FromImplicitCat(cat))

	b, err := json.Marshal(u)
	require.NoError(t, err)

	var decoded GetResponseRootOneOf200JSONResponseBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestResponseRootAnyOf_RoundTripDog(t *testing.T) {
	dog := ImplicitDog{Kind: Dog, Name: ptr("rex")}

	var u GetResponseRootAnyOf200JSONResponseBody
	require.NoError(t, u.FromImplicitDog(dog))

	b, err := json.Marshal(u)
	require.NoError(t, err)

	var decoded GetResponseRootAnyOf200JSONResponseBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitDog()
	require.NoError(t, err)
	require.Equal(t, dog, got)
}

func TestResponseItemsOneOf_RoundTripBothBranches(t *testing.T) {
	cat := ImplicitCat{Kind: Cat, Name: ptr("whiskers")}
	dog := ImplicitDog{Kind: Dog, Name: ptr("rex")}

	var catItem GetResponseItemsOneOf200JSONResponseBody_Items_Item
	require.NoError(t, catItem.FromImplicitCat(cat))
	var dogItem GetResponseItemsOneOf200JSONResponseBody_Items_Item
	require.NoError(t, dogItem.FromImplicitDog(dog))

	bCat, err := json.Marshal(catItem)
	require.NoError(t, err)
	bDog, err := json.Marshal(dogItem)
	require.NoError(t, err)

	var decodedCat, decodedDog GetResponseItemsOneOf200JSONResponseBody_Items_Item
	require.NoError(t, json.Unmarshal(bCat, &decodedCat))
	require.NoError(t, json.Unmarshal(bDog, &decodedDog))

	gotCat, err := decodedCat.AsImplicitCat()
	require.NoError(t, err)
	require.Equal(t, cat, gotCat)

	gotDog, err := decodedDog.AsImplicitDog()
	require.NoError(t, err)
	require.Equal(t, dog, gotDog)
}

func TestResponseDeepNested_RoundTrip(t *testing.T) {
	cat := ImplicitCat{Kind: Cat, Name: ptr("whiskers")}

	var inner GetResponseDeepNested200JSONResponseBody_Wrapper_Inner
	require.NoError(t, inner.FromImplicitCat(cat))

	b, err := json.Marshal(inner)
	require.NoError(t, err)

	var decoded GetResponseDeepNested200JSONResponseBody_Wrapper_Inner
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestBodyRootOneOf_RoundTripCat(t *testing.T) {
	cat := ImplicitCat{Kind: Cat, Name: ptr("whiskers")}

	var body PostBodyRootOneOfJSONBody
	require.NoError(t, body.FromImplicitCat(cat))

	b, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded PostBodyRootOneOfJSONBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestBodyPropertyOneOf_RoundTripDog(t *testing.T) {
	dog := ImplicitDog{Kind: Dog, Name: ptr("rex")}

	var pet PostBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, pet.FromImplicitDog(dog))

	b, err := json.Marshal(pet)
	require.NoError(t, err)

	var decoded PostBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitDog()
	require.NoError(t, err)
	require.Equal(t, dog, got)
}

func TestWebhookBodyRootOneOf_RoundTripCat(t *testing.T) {
	cat := ImplicitCat{Kind: Cat, Name: ptr("whiskers")}

	var body WebhookBodyRootOneOfJSONBody
	require.NoError(t, body.FromImplicitCat(cat))

	b, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded WebhookBodyRootOneOfJSONBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestWebhookBodyPropertyOneOf_RoundTripDog(t *testing.T) {
	dog := ImplicitDog{Kind: Dog, Name: ptr("rex")}

	var pet WebhookBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, pet.FromImplicitDog(dog))

	b, err := json.Marshal(pet)
	require.NoError(t, err)

	var decoded WebhookBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitDog()
	require.NoError(t, err)
	require.Equal(t, dog, got)
}

func TestCallbackBodyRootOneOf_RoundTripCat(t *testing.T) {
	cat := ImplicitCat{Kind: Cat, Name: ptr("whiskers")}

	var body CallbackBodyRootOneOfJSONBody
	require.NoError(t, body.FromImplicitCat(cat))

	b, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded CallbackBodyRootOneOfJSONBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestCallbackBodyPropertyOneOf_RoundTripDog(t *testing.T) {
	dog := ImplicitDog{Kind: Dog, Name: ptr("rex")}

	var pet CallbackBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, pet.FromImplicitDog(dog))

	b, err := json.Marshal(pet)
	require.NoError(t, err)

	var decoded CallbackBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsImplicitDog()
	require.NoError(t, err)
	require.Equal(t, dog, got)
}

func TestMergeOverwritesPriorBranch(t *testing.T) {
	cat := ImplicitCat{Kind: Cat, Name: ptr("whiskers")}
	dog := ImplicitDog{Kind: Dog, Name: ptr("rex")}

	var u GetResponseRootOneOf200JSONResponseBody
	require.NoError(t, u.FromImplicitCat(cat))
	require.NoError(t, u.MergeImplicitDog(dog))

	b, err := json.Marshal(u)
	require.NoError(t, err)

	var decoded GetResponseRootOneOf200JSONResponseBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	gotDog, err := decoded.AsImplicitDog()
	require.NoError(t, err)
	require.Equal(t, dog, gotDog)
}
