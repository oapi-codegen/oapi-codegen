package discriminator

import (
	"encoding/json"
	"fmt"
)

// ExamplePetDiscriminator demonstrates standard one-level inheritance discriminator usage.
func ExamplePetDiscriminator() {
	catJSON := `{"petType": "cat", "name": "Fluffy", "meow": true}`

	var pet Pet
	if err := json.Unmarshal([]byte(catJSON), &pet); err != nil {
		panic(err)
	}

	fmt.Printf("Pet: %s (type: %s)\n", pet.Name, pet.Discriminator())

	// IsCat is here only for clarity. Use only `value, err := pet.AsCat()` for concrete type.
	if pet.IsCat() {
		cat, _ := pet.AsCat()
		fmt.Printf("Cat meows: %v\n", *cat.Meow)
	}
}

// ExampleValueByDiscriminator demonstrates type detection using ValueByDiscriminator + switch type.
// Recommend using Is<Type> helpers for this.
func ExampleValueByDiscriminator() {
	catJSON := `{"petType": "cat", "name": "Whiskers", "meow": true}`

	var pet Pet
	if err := json.Unmarshal([]byte(catJSON), &pet); err != nil {
		panic(err)
	}

	value, err := pet.ValueByDiscriminator()
	if err != nil {
		panic(err)
	}

	switch v := value.(type) {
	case *Cat:
		fmt.Printf("This is a cat: %s (meows: %v)\n", v.Name, *v.Meow)
	case *Dog:
		fmt.Printf("This is a dog: %s (barks: %v)\n", v.Name, *v.Bark)
	default:
		fmt.Printf("Unknown pet type: %T\n", v)
	}
}

// ExampleNestedDiscriminators demonstrates multi-level discriminator hierarchy.
func ExampleNestedDiscriminators() {
	houseCatJSON := `{
		"animalType": "domestic",
		"domesticType": "housecat",
		"name": "Whiskers",
		"owner": "John",
		"indoor": true
	}`

	var animal Animal
	if err := json.Unmarshal([]byte(houseCatJSON), &animal); err != nil {
		panic(err)
	}

	fmt.Printf("Animal: %s (type: %s)\n", animal.Name, animal.Discriminator())

	// Navigate through hierarchy or use `value, err := As<Type>` without Is<Type>.
	// The hierarchy navigation is shown for clarity.
	if animal.IsDomesticAnimal() {
		domestic, _ := animal.AsDomesticAnimal()
		fmt.Printf("  Domestic type: %s, Owner: %s\n", domestic.Discriminator(), *domestic.Owner)

		// Further navigate to concrete type
		if domestic.IsHouseCat() {
			cat, _ := domestic.AsHouseCat()
			fmt.Printf("    House cat, Indoor: %v\n", *cat.Indoor)
		}
	}

	// Another example: Lion
	lionJSON := `{
		"animalType": "wild",
		"wildType": "lion",
		"name": "Simba",
		"habitat": "Savanna",
		"maneColor": "golden"
	}`

	var wildAnimal Animal
	if err := json.Unmarshal([]byte(lionJSON), &wildAnimal); err != nil {
		panic(err)
	}

	if wildAnimal.IsWildAnimal() {
		wild, _ := wildAnimal.AsWildAnimal()
		fmt.Printf("\nWild animal: %s, habitat: %s\n", wild.Name, *wild.Habitat)

		if wild.IsLion() {
			lion, _ := wild.AsLion()
			fmt.Printf("  Lion with mane color: %s\n", lion.ManeColor)
		}
	}
}

// ExampleProcessingArray demonstrates filtering arrays of polymorphic objects.
func ExampleProcessingArray() {
	animalsJSON := `[
		{"animalType": "domestic", "domesticType": "housecat", "name": "Whiskers", "owner": "Alice", "indoor": true},
		{"animalType": "wild", "wildType": "lion", "name": "Simba", "habitat": "Savanna"},
		{"animalType": "domestic", "domesticType": "housedog", "name": "Buddy", "owner": "Bob", "trained": true}
	]`

	var animals []Animal
	if err := json.Unmarshal([]byte(animalsJSON), &animals); err != nil {
		panic(err)
	}

	// Select cats
	var cats []*HouseCat
	for _, animal := range animals {
		if animal.IsDomesticAnimal() {
			domestic, _ := animal.AsDomesticAnimal()

			if domestic.IsHouseCat() {
				cat, _ := domestic.AsHouseCat()
				cats = append(cats, cat)
			}
		}
	}
	fmt.Printf("Found %d cats\n", len(cats))
}
