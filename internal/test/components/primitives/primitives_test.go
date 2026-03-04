package primitives

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPetTypeInstantiation(t *testing.T) {
	age := int32(5)
	weight := 12.5
	isGood := true

	pet := Pet{
		Name:   "Fido",
		Age:    &age,
		Weight: &weight,
		IsGood: &isGood,
		BornAt: openapi_types.Date{Time: time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)},
	}

	assert.Equal(t, "Fido", pet.Name)
	assert.Equal(t, int32(5), *pet.Age)
	assert.Equal(t, 12.5, *pet.Weight)
	assert.True(t, *pet.IsGood)
}

func TestPetJSONRoundTrip(t *testing.T) {
	input := `{"name":"Fido","age":5,"weight":12.5,"is_good":true,"born_at":"2020-01-15"}`

	var pet Pet
	err := json.Unmarshal([]byte(input), &pet)
	require.NoError(t, err)

	assert.Equal(t, "Fido", pet.Name)
	assert.Equal(t, int32(5), *pet.Age)
	assert.Equal(t, 12.5, *pet.Weight)
	assert.True(t, *pet.IsGood)
	assert.Equal(t, 2020, pet.BornAt.Year())
	assert.Equal(t, time.January, pet.BornAt.Month())
	assert.Equal(t, 15, pet.BornAt.Day())

	marshaled, err := json.Marshal(pet)
	require.NoError(t, err)

	var roundTripped Pet
	err = json.Unmarshal(marshaled, &roundTripped)
	require.NoError(t, err)
	assert.Equal(t, pet.Name, roundTripped.Name)
	assert.Equal(t, *pet.Age, *roundTripped.Age)
}

// From issue-579: aliased date types should unmarshal correctly.
func TestAliasedDateRoundTrip(t *testing.T) {
	input := `{"name":"Fido","born":"2022-05-19","born_at":"2022-05-20"}`

	var pet Pet
	err := json.Unmarshal([]byte(input), &pet)
	require.NoError(t, err)

	assert.Equal(t, 2022, pet.Born.Year())
	assert.Equal(t, time.May, pet.Born.Month())
	assert.Equal(t, 19, pet.Born.Day())

	assert.Equal(t, 2022, pet.BornAt.Year())
	assert.Equal(t, time.May, pet.BornAt.Month())
	assert.Equal(t, 20, pet.BornAt.Day())
}

func TestAllFormatsRoundTrip(t *testing.T) {
	af := AllFormats{
		DateField:     openapi_types.Date{Time: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)},
		DateTimeField: time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		UuidField:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		EmailField:    "test@example.com",
		Int32Field:    42,
		Int64Field:    9999999999,
		FloatField:    3.14,
		DoubleField:   2.718281828,
		ByteField:     []byte("aGVsbG8="),
	}

	data, err := json.Marshal(af)
	require.NoError(t, err)

	var roundTripped AllFormats
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	assert.Equal(t, af.Int32Field, roundTripped.Int32Field)
	assert.Equal(t, af.Int64Field, roundTripped.Int64Field)
	assert.InDelta(t, float64(af.FloatField), float64(roundTripped.FloatField), 0.001)
	assert.Equal(t, af.DoubleField, roundTripped.DoubleField)
	assert.Equal(t, af.EmailField, roundTripped.EmailField)
}

func TestStringEnumValues(t *testing.T) {
	assert.Equal(t, StringEnum("cat"), Cat)
	assert.Equal(t, StringEnum("dog"), Dog)
	assert.Equal(t, StringEnum("mouse"), Mouse)

	assert.True(t, Cat.Valid())
	assert.True(t, Dog.Valid())
	assert.True(t, Mouse.Valid())
	assert.False(t, StringEnum("fish").Valid())
}

func TestIntEnumValues(t *testing.T) {
	assert.Equal(t, IntEnum(1), IntEnumN1)
	assert.Equal(t, IntEnum(2), IntEnumN2)
	assert.Equal(t, IntEnum(3), IntEnumN3)

	assert.True(t, IntEnumN1.Valid())
	assert.True(t, IntEnumN2.Valid())
	assert.True(t, IntEnumN3.Valid())
	assert.False(t, IntEnum(99).Valid())
}

// From issue-illegal_enum_names: enum values with invalid Go identifiers
// must be sanitized into valid const names.
func TestEdgeCaseEnumNames(t *testing.T) {
	// The fact that these compile proves the names are valid Go identifiers.
	// We verify the string values are correct.
	assert.Equal(t, EdgeCaseEnum(""), EdgeCaseEnumEmpty)
	assert.Equal(t, EdgeCaseEnum("Foo"), EdgeCaseEnumFoo)
	assert.Equal(t, EdgeCaseEnum("Bar"), EdgeCaseEnumBar)
	assert.Equal(t, EdgeCaseEnum("Foo Bar"), EdgeCaseEnumFooBar)
	assert.Equal(t, EdgeCaseEnum("Foo-Bar"), EdgeCaseEnumFooBar1)
	assert.Equal(t, EdgeCaseEnum("1Foo"), EdgeCaseEnumN1Foo)
	assert.Equal(t, EdgeCaseEnum(" Foo"), EdgeCaseEnumFoo1)
	assert.Equal(t, EdgeCaseEnum(" Foo "), EdgeCaseEnumFoo2)
	assert.Equal(t, EdgeCaseEnum("_Foo_"), EdgeCaseEnumUnderscoreFoo)
	assert.Equal(t, EdgeCaseEnum("1"), EdgeCaseEnumN1)

	// All edge case values should be valid.
	assert.True(t, EdgeCaseEnumEmpty.Valid())
	assert.True(t, EdgeCaseEnumFooBar.Valid())
	assert.True(t, EdgeCaseEnumN1Foo.Valid())
}

func TestEnumJSONRoundTrip(t *testing.T) {
	type wrapper struct {
		S StringEnum `json:"s"`
		I IntEnum    `json:"i"`
		E EdgeCaseEnum `json:"e"`
	}

	w := wrapper{S: Dog, I: IntEnumN2, E: EdgeCaseEnumFooBar}
	data, err := json.Marshal(w)
	require.NoError(t, err)

	var roundTripped wrapper
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	assert.Equal(t, w, roundTripped)
}

func TestCustomFormatString(t *testing.T) {
	// CustomFormatString is a type alias for string.
	var s CustomFormatString = "hello"
	assert.Equal(t, "hello", s)
}
