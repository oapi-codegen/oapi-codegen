package anyof

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnyOfAnimalFromCat(t *testing.T) {
	cat := Cat{Type: "cat", Name: "Whiskers", Indoor: boolPtr(true)}

	var animal AnyOfAnimal
	err := animal.FromCat(cat)
	require.NoError(t, err)

	data, err := json.Marshal(animal)
	require.NoError(t, err)

	var roundTripped AnyOfAnimal
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	gotCat, err := roundTripped.AsCat()
	require.NoError(t, err)
	assert.Equal(t, "Whiskers", gotCat.Name)
	assert.Equal(t, "cat", gotCat.Type)
	assert.True(t, *gotCat.Indoor)
}

func TestAnyOfAnimalFromDog(t *testing.T) {
	breed := "Golden Retriever"
	dog := Dog{Type: "dog", Name: "Buddy", Breed: &breed}

	var animal AnyOfAnimal
	err := animal.FromDog(dog)
	require.NoError(t, err)

	data, err := json.Marshal(animal)
	require.NoError(t, err)

	var roundTripped AnyOfAnimal
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	gotDog, err := roundTripped.AsDog()
	require.NoError(t, err)
	assert.Equal(t, "Buddy", gotDog.Name)
	assert.Equal(t, "Golden Retriever", *gotDog.Breed)
}

func TestSimpleAnyOfWithObjectVariant(t *testing.T) {
	name := "test"
	obj := SimpleAnyOf0{Name: &name}

	var sa SimpleAnyOf
	err := sa.FromSimpleAnyOf0(obj)
	require.NoError(t, err)

	data, err := json.Marshal(sa)
	require.NoError(t, err)

	var roundTripped SimpleAnyOf
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	got, err := roundTripped.AsSimpleAnyOf0()
	require.NoError(t, err)
	assert.Equal(t, "test", *got.Name)
}

func TestSimpleAnyOfWithIdVariant(t *testing.T) {
	id := 42
	obj := SimpleAnyOf1{Id: &id}

	var sa SimpleAnyOf
	err := sa.FromSimpleAnyOf1(obj)
	require.NoError(t, err)

	data, err := json.Marshal(sa)
	require.NoError(t, err)

	var roundTripped SimpleAnyOf
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	got, err := roundTripped.AsSimpleAnyOf1()
	require.NoError(t, err)
	assert.Equal(t, 42, *got.Id)
}

// From issue-1189: anyOf with duplicate/overlapping schema types.
func TestDuplicateAnyOfWithPlainString(t *testing.T) {
	var da DuplicateAnyOf
	err := da.FromDuplicateAnyOf0("hello")
	require.NoError(t, err)

	data, err := json.Marshal(da)
	require.NoError(t, err)
	assert.Equal(t, `"hello"`, string(data))
}

func TestDuplicateAnyOfWithEnumString(t *testing.T) {
	var da DuplicateAnyOf
	err := da.FromDuplicateAnyOf1(Alpha)
	require.NoError(t, err)

	data, err := json.Marshal(da)
	require.NoError(t, err)
	assert.Equal(t, `"alpha"`, string(data))

	// The enum value should be valid.
	assert.True(t, Alpha.Valid())
	assert.True(t, Beta.Valid())
	assert.True(t, Gamma.Valid())
}

func boolPtr(b bool) *bool {
	return &b
}
