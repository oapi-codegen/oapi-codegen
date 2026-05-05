package anonymous_inner_hoisting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func TestResponseRootOneOf_RoundTripCat(t *testing.T) {
	cat := Cat{Kind: CatKindCat, Name: ptr("whiskers")}

	var u GetResponseRootOneOf200JSONResponseBody
	require.NoError(t, u.FromCat(cat))

	b, err := json.Marshal(u)
	require.NoError(t, err)

	var decoded GetResponseRootOneOf200JSONResponseBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestResponseRootAnyOf_RoundTripDog(t *testing.T) {
	dog := Dog{Kind: DogKindDog, Name: ptr("rex")}

	var u GetResponseRootAnyOf200JSONResponseBody
	require.NoError(t, u.FromDog(dog))

	b, err := json.Marshal(u)
	require.NoError(t, err)

	var decoded GetResponseRootAnyOf200JSONResponseBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsDog()
	require.NoError(t, err)
	require.Equal(t, dog, got)
}

func TestResponseItemsOneOf_RoundTripBothBranches(t *testing.T) {
	cat := Cat{Kind: CatKindCat, Name: ptr("whiskers")}
	dog := Dog{Kind: DogKindDog, Name: ptr("rex")}

	var catItem GetResponseItemsOneOf200JSONResponseBody_Items_Item
	require.NoError(t, catItem.FromCat(cat))
	var dogItem GetResponseItemsOneOf200JSONResponseBody_Items_Item
	require.NoError(t, dogItem.FromDog(dog))

	bCat, err := json.Marshal(catItem)
	require.NoError(t, err)
	bDog, err := json.Marshal(dogItem)
	require.NoError(t, err)

	var decodedCat, decodedDog GetResponseItemsOneOf200JSONResponseBody_Items_Item
	require.NoError(t, json.Unmarshal(bCat, &decodedCat))
	require.NoError(t, json.Unmarshal(bDog, &decodedDog))

	gotCat, err := decodedCat.AsCat()
	require.NoError(t, err)
	require.Equal(t, cat, gotCat)

	gotDog, err := decodedDog.AsDog()
	require.NoError(t, err)
	require.Equal(t, dog, gotDog)
}

func TestResponseDeepNested_RoundTrip(t *testing.T) {
	cat := Cat{Kind: CatKindCat, Name: ptr("whiskers")}

	var inner GetResponseDeepNested200JSONResponseBody_Wrapper_Inner
	require.NoError(t, inner.FromCat(cat))

	b, err := json.Marshal(inner)
	require.NoError(t, err)

	var decoded GetResponseDeepNested200JSONResponseBody_Wrapper_Inner
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestBodyRootOneOf_RoundTripCat(t *testing.T) {
	cat := Cat{Kind: CatKindCat, Name: ptr("whiskers")}

	var body PostBodyRootOneOfJSONBody
	require.NoError(t, body.FromCat(cat))

	b, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded PostBodyRootOneOfJSONBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestBodyPropertyOneOf_RoundTripDog(t *testing.T) {
	dog := Dog{Kind: DogKindDog, Name: ptr("rex")}

	var pet PostBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, pet.FromDog(dog))

	b, err := json.Marshal(pet)
	require.NoError(t, err)

	var decoded PostBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsDog()
	require.NoError(t, err)
	require.Equal(t, dog, got)
}

func TestWebhookBodyRootOneOf_RoundTripCat(t *testing.T) {
	cat := Cat{Kind: CatKindCat, Name: ptr("whiskers")}

	var body WebhookBodyRootOneOfJSONBody
	require.NoError(t, body.FromCat(cat))

	b, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded WebhookBodyRootOneOfJSONBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestWebhookBodyPropertyOneOf_RoundTripDog(t *testing.T) {
	dog := Dog{Kind: DogKindDog, Name: ptr("rex")}

	var pet WebhookBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, pet.FromDog(dog))

	b, err := json.Marshal(pet)
	require.NoError(t, err)

	var decoded WebhookBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsDog()
	require.NoError(t, err)
	require.Equal(t, dog, got)
}

func TestCallbackBodyRootOneOf_RoundTripCat(t *testing.T) {
	cat := Cat{Kind: CatKindCat, Name: ptr("whiskers")}

	var body CallbackBodyRootOneOfJSONBody
	require.NoError(t, body.FromCat(cat))

	b, err := json.Marshal(body)
	require.NoError(t, err)

	var decoded CallbackBodyRootOneOfJSONBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsCat()
	require.NoError(t, err)
	require.Equal(t, cat, got)
}

func TestCallbackBodyPropertyOneOf_RoundTripDog(t *testing.T) {
	dog := Dog{Kind: DogKindDog, Name: ptr("rex")}

	var pet CallbackBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, pet.FromDog(dog))

	b, err := json.Marshal(pet)
	require.NoError(t, err)

	var decoded CallbackBodyPropertyOneOfJSONBody_Pet
	require.NoError(t, json.Unmarshal(b, &decoded))

	got, err := decoded.AsDog()
	require.NoError(t, err)
	require.Equal(t, dog, got)
}

func TestMergeOverwritesPriorBranch(t *testing.T) {
	cat := Cat{Kind: CatKindCat, Name: ptr("whiskers")}
	dog := Dog{Kind: DogKindDog, Name: ptr("rex")}

	var u GetResponseRootOneOf200JSONResponseBody
	require.NoError(t, u.FromCat(cat))
	require.NoError(t, u.MergeDog(dog))

	b, err := json.Marshal(u)
	require.NoError(t, err)

	var decoded GetResponseRootOneOf200JSONResponseBody
	require.NoError(t, json.Unmarshal(b, &decoded))

	gotDog, err := decoded.AsDog()
	require.NoError(t, err)
	require.Equal(t, dog, gotDog)
}
