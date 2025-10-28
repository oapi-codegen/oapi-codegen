package discriminator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test basic discriminator functionality
// All scenarios (standard, schema-level, inline) have same behavior.
func TestBasicDiscriminator(t *testing.T) {
	tests := []struct {
		name string
		json string
		test func(t *testing.T)
	}{
		{
			name: "Pet cat (standard one-level inheritance discriminator)",
			json: `{"petType":"cat","name":"Fluffy","meow":true}`,
			test: func(t *testing.T) {
				var pet Pet
				assert.NoError(t, json.Unmarshal([]byte(`{"petType":"cat","name":"Fluffy","meow":true}`), &pet))
				assert.Equal(t, "cat", pet.Discriminator())
				assert.True(t, pet.IsCat())

				cat, err := pet.AsCat()
				assert.NoError(t, err)
				assert.True(t, *cat.Meow)

				// Wrong type conversion should fail
				_, err = pet.AsDog()
				assert.Error(t, err)
			},
		},
		{
			name: "PetActivity (discriminator at schema level with allOf)",
			json: `{"activityType":"feeding","duration":15,"foodType":"kibble"}`,
			test: func(t *testing.T) {
				var activity PetActivity
				assert.NoError(t, json.Unmarshal([]byte(`{"activityType":"feeding","duration":15,"foodType":"kibble"}`), &activity))
				assert.Equal(t, "feeding", activity.Discriminator())
				assert.True(t, activity.IsFeedingActivity())

				feeding, err := activity.AsFeedingActivity()
				assert.NoError(t, err)
				assert.Equal(t, "kibble", feeding.FoodType)
			},
		},
		{
			name: "HealthRecord (discriminator in inline allOf element)",
			json: `{"recordType":"vaccination","date":"2024-01-15","vaccine":"Rabies"}`,
			test: func(t *testing.T) {
				var record HealthRecord
				assert.NoError(t, json.Unmarshal([]byte(`{"recordType":"vaccination","date":"2024-01-15","vaccine":"Rabies"}`), &record))
				assert.Equal(t, "vaccination", record.Discriminator())
				assert.True(t, record.IsVaccinationRecord())

				vaccination, err := record.AsVaccinationRecord()
				assert.NoError(t, err)
				assert.Equal(t, "Rabies", vaccination.Vaccine)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// Test ValueByDiscriminator method.
func TestValueByDiscriminator(t *testing.T) {
	catJSON := `{"petType":"cat","name":"Fluffy","meow":true}`

	var pet Pet
	assert.NoError(t, json.Unmarshal([]byte(catJSON), &pet))

	value, err := pet.ValueByDiscriminator()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	cat, ok := value.(*Cat)
	assert.True(t, ok, "Expected *Cat type")
	assert.Equal(t, "cat", cat.PetType)
	assert.True(t, *cat.Meow)
}

// Test nested discriminators (multi-level hierarchy).
func TestNestedDiscriminators(t *testing.T) {
	houseCatJSON := `{
		"animalType":"domestic",
		"domesticType":"housecat",
		"name":"Whiskers",
		"owner":"Alice",
		"indoor":true
	}`

	var animal Animal
	assert.NoError(t, json.Unmarshal([]byte(houseCatJSON), &animal))
	assert.Equal(t, "domestic", animal.Discriminator())

	// Level 1 -> Level 2
	assert.True(t, animal.IsDomesticAnimal())
	domestic, err := animal.AsDomesticAnimal()
	assert.NoError(t, err)
	assert.Equal(t, "housecat", domestic.Discriminator())

	// Level 2 -> Level 3
	assert.True(t, domestic.IsHouseCat())
	cat, err := domestic.AsHouseCat()
	assert.NoError(t, err)
	assert.True(t, *cat.Indoor)
}

// Test array processing pattern.
func TestArrayProcessing(t *testing.T) {
	animalsJSON := `[
		{"animalType":"domestic","domesticType":"housecat","name":"Cat1","owner":"Owner1"},
		{"animalType":"wild","wildType":"lion","name":"Lion1","habitat":"Savanna"},
		{"animalType":"domestic","domesticType":"housedog","name":"Dog1","owner":"Owner2"}
	]`

	var animals []Animal
	assert.NoError(t, json.Unmarshal([]byte(animalsJSON), &animals))
	assert.Len(t, animals, 3)

	// Filter by animalType and domesticType
	var cats []*HouseCat
	for _, animal := range animals {
		if animal.IsDomesticAnimal() {
			domestic, _ := animal.AsDomesticAnimal()
			if domestic.IsHouseCat() {
				cat, err := domestic.AsHouseCat()
				assert.NoError(t, err)
				cats = append(cats, cat)
			}
		}
	}
	assert.Len(t, cats, 1)
	assert.Equal(t, "Cat1", cats[0].Name)
}
