// Package param provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package param

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/oapi-codegen/runtime"
)

// Defines values for CatBreed.
const (
	CatBreedMaineCoon CatBreed = "maine_coon"
	CatBreedRagdoll   CatBreed = "ragdoll"
)

// Defines values for ComplexPetBreed.
const (
	ComplexPetBreedBeagle    ComplexPetBreed = "beagle"
	ComplexPetBreedLizard    ComplexPetBreed = "lizard"
	ComplexPetBreedMaineCoon ComplexPetBreed = "maine_coon"
	ComplexPetBreedPoodle    ComplexPetBreed = "poodle"
	ComplexPetBreedRagdoll   ComplexPetBreed = "ragdoll"
)

// Defines values for DogBreed.
const (
	Beagle DogBreed = "beagle"
	Poodle DogBreed = "poodle"
)

// Cat defines model for cat.
type Cat struct {
	Breed CatBreed `json:"breed"`
}

// CatBreed defines model for Cat.Breed.
type CatBreed string

// ComplexPet defines model for complexPet.
type ComplexPet struct {
	Breed ComplexPetBreed `json:"breed"`
	union json.RawMessage
}

// ComplexPetBreed defines model for ComplexPet.Breed.
type ComplexPetBreed string

// Dog defines model for dog.
type Dog struct {
	Breed DogBreed `json:"breed"`
}

// DogBreed defines model for Dog.Breed.
type DogBreed string

// Pet defines model for pet.
type Pet struct {
	union json.RawMessage
}

// AsDog returns the union data inside the ComplexPet as a Dog
func (t ComplexPet) AsDog() (Dog, error) {
	var body Dog
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *ComplexPet) prepareDog(v Dog) ([]byte, error) {
	t.Breed = (ComplexPetBreed)(string(v.Breed))
	return json.Marshal(v)
}

// FromDog overwrites any union data inside the ComplexPet as the provided Dog
func (t *ComplexPet) FromDog(v Dog) error {
	b, err := t.prepareDog(v)
	t.union = b
	return err
}

// MergeDog performs a merge with any union data inside the ComplexPet, using the provided Dog
func (t *ComplexPet) MergeDog(v Dog) error {
	b, err := t.prepareDog(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

// AsCat returns the union data inside the ComplexPet as a Cat
func (t ComplexPet) AsCat() (Cat, error) {
	var body Cat
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *ComplexPet) prepareCat(v Cat) ([]byte, error) {
	t.Breed = (ComplexPetBreed)(string(v.Breed))
	return json.Marshal(v)
}

// FromCat overwrites any union data inside the ComplexPet as the provided Cat
func (t *ComplexPet) FromCat(v Cat) error {
	b, err := t.prepareCat(v)
	t.union = b
	return err
}

// MergeCat performs a merge with any union data inside the ComplexPet, using the provided Cat
func (t *ComplexPet) MergeCat(v Cat) error {
	b, err := t.prepareCat(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

func (t ComplexPet) Discriminator() (string, error) {
	var discriminator struct {
		Discriminator string `json:"breed"`
	}
	err := json.Unmarshal(t.union, &discriminator)
	return discriminator.Discriminator, err
}

func (t ComplexPet) ValueByDiscriminator() (interface{}, error) {
	discriminator, err := t.Discriminator()
	if err != nil {
		return nil, err
	}
	switch discriminator {
	case "beagle":
		return t.AsDog()
	case "maine_coon":
		return t.AsCat()
	case "poodle":
		return t.AsDog()
	case "ragdoll":
		return t.AsCat()
	default:
		return nil, errors.New("unknown discriminator value: " + discriminator)
	}
}

func (t ComplexPet) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	if err != nil {
		return nil, err
	}
	object := make(map[string]json.RawMessage)
	if t.union != nil {
		err = json.Unmarshal(b, &object)
		if err != nil {
			return nil, err
		}
	}

	object["breed"], err = json.Marshal(t.Breed)
	if err != nil {
		return nil, fmt.Errorf("error marshaling 'breed': %w", err)
	}

	b, err = json.Marshal(object)
	return b, err
}

func (t *ComplexPet) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	object := make(map[string]json.RawMessage)
	err = json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if raw, found := object["breed"]; found {
		err = json.Unmarshal(raw, &t.Breed)
		if err != nil {
			return fmt.Errorf("error reading 'breed': %w", err)
		}
	}

	return err
}

// AsDog returns the union data inside the Pet as a Dog
func (t Pet) AsDog() (Dog, error) {
	var body Dog
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *Pet) prepareDog(v Dog) ([]byte, error) {
	return json.Marshal(v)
}

// FromDog overwrites any union data inside the Pet as the provided Dog
func (t *Pet) FromDog(v Dog) error {
	b, err := t.prepareDog(v)
	t.union = b
	return err
}

// MergeDog performs a merge with any union data inside the Pet, using the provided Dog
func (t *Pet) MergeDog(v Dog) error {
	b, err := t.prepareDog(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

// AsCat returns the union data inside the Pet as a Cat
func (t Pet) AsCat() (Cat, error) {
	var body Cat
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *Pet) prepareCat(v Cat) ([]byte, error) {
	return json.Marshal(v)
}

// FromCat overwrites any union data inside the Pet as the provided Cat
func (t *Pet) FromCat(v Cat) error {
	b, err := t.prepareCat(v)
	t.union = b
	return err
}

// MergeCat performs a merge with any union data inside the Pet, using the provided Cat
func (t *Pet) MergeCat(v Cat) error {
	b, err := t.prepareCat(v)
	if err != nil {
		return err
	}

	merged, err := runtime.JSONMerge(t.union, b)
	t.union = merged
	return err
}

func (t Pet) Discriminator() (string, error) {
	var discriminator struct {
		Discriminator string `json:"breed"`
	}
	err := json.Unmarshal(t.union, &discriminator)
	return discriminator.Discriminator, err
}

func (t Pet) ValueByDiscriminator() (interface{}, error) {
	discriminator, err := t.Discriminator()
	if err != nil {
		return nil, err
	}
	switch discriminator {
	case "beagle":
		return t.AsDog()
	case "maine_coon":
		return t.AsCat()
	case "poodle":
		return t.AsDog()
	case "ragdoll":
		return t.AsCat()
	default:
		return nil, errors.New("unknown discriminator value: " + discriminator)
	}
}

func (t Pet) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *Pet) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}
