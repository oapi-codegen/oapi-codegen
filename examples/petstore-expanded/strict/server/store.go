package server

import (
	"fmt"
	"sync"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/strict/api"
)

// ErrNotFound is returned when a pet is not found in the store.
var ErrNotFound = fmt.Errorf("not found")

// Store implements a simple in-memory pet store.
type Store struct {
	Pets   map[int64]api.Pet
	NextId int64
	Lock   sync.Mutex
}

// NewStore creates a new Store with an empty pet map.
func NewStore() *Store {
	return &Store{
		Pets:   make(map[int64]api.Pet),
		NextId: 1000,
	}
}

// FindPets returns all pets, optionally filtered by tags and limited in count.
func (p *Store) FindPets(tags *[]string, limit *int32) []api.Pet {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var result []api.Pet

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
func (p *Store) AddPet(newPet api.NewPet) api.Pet {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var pet api.Pet
	pet.Name = newPet.Name
	pet.Tag = newPet.Tag
	pet.Id = p.NextId
	p.NextId++

	p.Pets[pet.Id] = pet

	return pet
}

// FindPetByID returns a pet by its ID, or ErrNotFound if it doesn't exist.
func (p *Store) FindPetByID(id int64) (api.Pet, error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[id]
	if !found {
		return api.Pet{}, fmt.Errorf("could not find pet with ID %d: %w", id, ErrNotFound)
	}

	return pet, nil
}

// DeletePet deletes a pet by its ID, or returns ErrNotFound if it doesn't exist.
func (p *Store) DeletePet(id int64) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	_, found := p.Pets[id]
	if !found {
		return fmt.Errorf("could not find pet with ID %d: %w", id, ErrNotFound)
	}
	delete(p.Pets, id)

	return nil
}
