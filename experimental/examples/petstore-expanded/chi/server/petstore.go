//go:build go1.22

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	petstore "github.com/oapi-codegen/oapi-codegen/experimental/examples/petstore-expanded"
)

// PetStore implements the ServerInterface.
type PetStore struct {
	Pets   map[int64]petstore.Pet
	NextId int64
	Lock   sync.Mutex
}

// Make sure we conform to ServerInterface
var _ ServerInterface = (*PetStore)(nil)

// NewPetStore creates a new PetStore.
func NewPetStore() *PetStore {
	return &PetStore{
		Pets:   make(map[int64]petstore.Pet),
		NextId: 1000,
	}
}

// sendPetStoreError wraps sending of an error in the Error format.
func sendPetStoreError(w http.ResponseWriter, code int, message string) {
	petErr := petstore.Error{
		Code:    int32(code),
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(petErr)
}

// FindPets returns all pets, optionally filtered by tags and limited.
func (p *PetStore) FindPets(w http.ResponseWriter, r *http.Request, params FindPetsParams) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var result []petstore.Pet

	for _, pet := range p.Pets {
		if params.Tags != nil {
			// If we have tags, filter pets by tag
			for _, t := range *params.Tags {
				if pet.Tag != nil && (*pet.Tag == t) {
					result = append(result, pet)
				}
			}
		} else {
			// Add all pets if we're not filtering
			result = append(result, pet)
		}

		if params.Limit != nil {
			l := int(*params.Limit)
			if len(result) >= l {
				// We're at the limit
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// AddPet creates a new pet.
func (p *PetStore) AddPet(w http.ResponseWriter, r *http.Request) {
	// We expect a NewPet object in the request body.
	var newPet petstore.NewPet
	if err := json.NewDecoder(r.Body).Decode(&newPet); err != nil {
		sendPetStoreError(w, http.StatusBadRequest, "Invalid format for NewPet")
		return
	}

	// We now have a pet, let's add it to our "database".
	p.Lock.Lock()
	defer p.Lock.Unlock()

	// We handle pets, not NewPets, which have an additional ID field
	var pet petstore.Pet
	pet.Name = newPet.Name
	pet.Tag = newPet.Tag
	pet.ID = p.NextId
	p.NextId++

	// Insert into map
	p.Pets[pet.ID] = pet

	// Now, we have to return the Pet
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(pet)
}

// FindPetByID returns a pet by ID.
func (p *PetStore) FindPetByID(w http.ResponseWriter, r *http.Request, id int64) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[id]
	if !found {
		sendPetStoreError(w, http.StatusNotFound, fmt.Sprintf("Could not find pet with ID %d", id))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(pet)
}

// DeletePet deletes a pet by ID.
func (p *PetStore) DeletePet(w http.ResponseWriter, r *http.Request, id int64) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	_, found := p.Pets[id]
	if !found {
		sendPetStoreError(w, http.StatusNotFound, fmt.Sprintf("Could not find pet with ID %d", id))
		return
	}
	delete(p.Pets, id)

	w.WriteHeader(http.StatusNoContent)
}
