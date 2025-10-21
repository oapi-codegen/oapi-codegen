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
		test func(t *testing.T, jsonData string)
	}{
		{
			name: "Pet cat (standard one-level inheritance discriminator)",
			json: `{"petType":"cat","name":"Fluffy","meow":true}`,
			test: func(t *testing.T, jsonData string) {
				var pet Pet
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &pet))
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
			test: func(t *testing.T, jsonData string) {
				var activity PetActivity
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &activity))
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
			test: func(t *testing.T, jsonData string) {
				var record HealthRecord
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &record))
				assert.Equal(t, "vaccination", record.Discriminator())
				assert.True(t, record.IsVaccinationRecord())

				vaccination, err := record.AsVaccinationRecord()
				assert.NoError(t, err)
				assert.Equal(t, "Rabies", vaccination.Vaccine)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, tt.json)
		})
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

// TestArrayProcessing demonstrates array processing and filtering with discriminators.
func TestArrayProcessing(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		testFunc func(t *testing.T, jsonData string)
	}{
		{
			name: "Animal array filtering (nested discriminators)",
			jsonData: `[
				{"animalType":"domestic","domesticType":"housecat","name":"Cat1","owner":"Owner1"},
				{"animalType":"wild","wildType":"lion","name":"Lion1","habitat":"Savanna"},
				{"animalType":"domestic","domesticType":"housedog","name":"Dog1","owner":"Owner2"}
			]`,
			testFunc: func(t *testing.T, jsonData string) {
				var animals []Animal
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &animals))
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
			},
		},
		{
			name: "Pet array filtering",
			jsonData: `[
				{"petType":"cat","name":"Fluffy","meow":true},
				{"petType":"dog","name":"Buddy","bark":true},
				{"petType":"mouse","name":"Jerry","pestType":"mouse","habitat":"house","tailLength":5.5},
				{"petType":"cat","name":"Whiskers","meow":false}
			]`,
			testFunc: func(t *testing.T, jsonData string) {
				var pets []Pet
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &pets))
				assert.Len(t, pets, 4)

				// Filter Mouse objects
				var mice []*Mouse
				for _, pet := range pets {
					if pet.IsMouse() {
						mouse, err := pet.AsMouse()
						assert.NoError(t, err)
						mice = append(mice, mouse)
					}
				}
				assert.Len(t, mice, 1)
				assert.Equal(t, "Jerry", mice[0].Name)
				assert.Equal(t, "mouse", mice[0].PetType)
				assert.Equal(t, "mouse", mice[0].PestType)

				// Filter Cat objects
				var cats []*Cat
				for _, pet := range pets {
					if pet.IsCat() {
						cat, _ := pet.AsCat()
						cats = append(cats, cat)
					}
				}
				assert.Len(t, cats, 2)
				assert.Equal(t, "Fluffy", cats[0].Name)
			},
		},
		{
			name: "Pest array filtering",
			// Note: Currently Mouse is the only Pest type, so this array contains only mice.
			// When additional Pest types are added (e.g., rat, cockroach), this test should
			// be updated to demonstrate filtering different types, similar to Pet array filtering.
			jsonData: `[
				{"pestType":"mouse","name":"Jerry","habitat":"house","tailLength":5.5},
				{"pestType":"mouse","name":"Stuart","habitat":"apartment","tailLength":4.0},
				{"pestType":"mouse","name":"Mickey","habitat":"cartoon","tailLength":6.0}
			]`,
			testFunc: func(t *testing.T, jsonData string) {
				var pests []Pest
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &pests))
				assert.Len(t, pests, 3)

				for _, pest := range pests {
					_, err := pest.AsMouse()
					assert.NoError(t, err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t, tt.jsonData)
		})
	}
}

// TestMouseMultipleDiscriminators tests Mouse type combining Pet and Pest discriminators.
func TestMouseMultipleDiscriminators(t *testing.T) {
	tests := []struct {
		name string
		json string
		test func(t *testing.T, jsonData string)
	}{
		{
			name: "Mouse with both discriminator properties",
			json: `{"petType":"mouse","name":"Jerry","pestType":"mouse","habitat":"house","tailLength":5.5}`,
			test: func(t *testing.T, jsonData string) {
				var mouse Mouse
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &mouse))

				// Verify all fields are present
				assert.Equal(t, "mouse", mouse.PetType, "petType from Pet")
				assert.Equal(t, "Jerry", mouse.Name, "name from Pet")
				assert.Equal(t, "mouse", mouse.PestType, "pestType from Pest")
				assert.Equal(t, "house", mouse.Habitat, "habitat from Pest")
				assert.Equal(t, float32(5.5), *mouse.TailLength, "tailLength from Mouse")
			},
		},
		{
			name: "Mouse conversion via Pet and Pest interfaces",
			json: `{"petType":"mouse","name":"Mickey","pestType":"mouse","habitat":"house","tailLength":4.5}`,
			test: func(t *testing.T, jsonData string) {
				// Convert via Pet interface
				var pet Pet
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &pet))
				assert.Equal(t, "mouse", pet.Discriminator())
				assert.True(t, pet.IsMouse())

				mouseFromPet, err := pet.AsMouse()
				assert.NoError(t, err)
				assert.Equal(t, "Mickey", mouseFromPet.Name)
				assert.Equal(t, "mouse", mouseFromPet.PetType)
				assert.Equal(t, "mouse", mouseFromPet.PestType)
				assert.Equal(t, "house", mouseFromPet.Habitat)
				assert.Equal(t, float32(4.5), *mouseFromPet.TailLength)

				// Convert via Pest interface (same JSON)
				var pest Pest
				assert.NoError(t, json.Unmarshal([]byte(jsonData), &pest))
				assert.Equal(t, "mouse", pest.Discriminator())
				assert.True(t, pest.IsMouse())

				mouseFromPest, err := pest.AsMouse()
				assert.NoError(t, err)

				// Verify both conversions yield identical Mouse objects
				assert.Equal(t, mouseFromPet, mouseFromPest)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, tt.json)
		})
	}
}
