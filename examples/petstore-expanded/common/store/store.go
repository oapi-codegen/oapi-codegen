package store

import (
	"fmt"
	"sync"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/models"
)

// ErrNotFound is returned when a pet is not found in the store.
var ErrNotFound = fmt.Errorf("not found")

// PetStore implements a simple in-memory pet store.
type PetStore struct {
	Pets   map[int64]models.Pet
	NextId int64
	Lock   sync.Mutex
}

// NewPetStore creates a new PetStore with an empty pet map.
func NewPetStore() *PetStore {
	return &PetStore{
		Pets:   make(map[int64]models.Pet),
		NextId: 1000,
	}
}

// FindPets returns all pets, optionally filtered by tags and limited in count.
func (p *PetStore) FindPets(tags *[]string, limit *int32) []models.Pet {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var result []models.Pet

	for _, pet := range p.Pets {
		if tags != nil {
			for _, t := range *tags {
				if pet.Tag != nil && (*pet.Tag == t) {
					result = append(result, pet)
				}
			}
		} else {
			result = append(result, pet)
		}

		if limit != nil {
			l := int(*limit)
			if len(result) >= l {
				break
			}
		}
	}

	return result
}

// AddPet adds a new pet to the store and returns the created pet with its assigned ID.
func (p *PetStore) AddPet(newPet models.NewPet) models.Pet {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var pet models.Pet
	pet.Name = newPet.Name
	pet.Tag = newPet.Tag
	pet.Id = p.NextId
	p.NextId++

	p.Pets[pet.Id] = pet

	return pet
}

// FindPetByID returns a pet by its ID, or ErrNotFound if it doesn't exist.
func (p *PetStore) FindPetByID(id int64) (models.Pet, error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[id]
	if !found {
		return models.Pet{}, fmt.Errorf("could not find pet with ID %d: %w", id, ErrNotFound)
	}

	return pet, nil
}

// DeletePet deletes a pet by its ID, or returns ErrNotFound if it doesn't exist.
func (p *PetStore) DeletePet(id int64) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	_, found := p.Pets[id]
	if !found {
		return fmt.Errorf("could not find pet with ID %d: %w", id, ErrNotFound)
	}
	delete(p.Pets, id)

	return nil
}
