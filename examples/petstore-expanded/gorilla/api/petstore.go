//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../petstore-expanded.yaml

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type PetStore struct {
	Pets   map[int64]Pet
	NextId int64
	Lock   sync.Mutex
}

// Make sure we conform to ServerInterface

var _ ServerInterface = (*PetStore)(nil)

func NewPetStore() *PetStore {
	return &PetStore{
		Pets:   make(map[int64]Pet),
		NextId: 1000,
	}
}

// sendPetStoreError wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendPetStoreError(w http.ResponseWriter, code int, message string) {
	petErr := Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(petErr)
}

// FindPets implements all the handlers in the ServerInterface
func (p *PetStore) FindPets(w http.ResponseWriter, r *http.Request, params FindPetsParams) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var result []Pet

	for _, pet := range p.Pets {
		if params.Tags != nil {
			// If we have tags,  filter pets by tag
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

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func (p *PetStore) AddPet(w http.ResponseWriter, r *http.Request) {
	// We expect a NewPet object in the request body.
	var newPet NewPet
	if err := json.NewDecoder(r.Body).Decode(&newPet); err != nil {
		sendPetStoreError(w, http.StatusBadRequest, "Invalid format for NewPet")
		return
	}

	// We now have a pet, let's add it to our "database".

	// We're always asynchronous, so lock unsafe operations below
	p.Lock.Lock()
	defer p.Lock.Unlock()

	// We handle pets, not NewPets, which have an additional ID field
	var pet Pet
	pet.Name = newPet.Name
	pet.Tag = newPet.Tag
	pet.Id = p.NextId
	p.NextId++

	// Insert into map
	p.Pets[pet.Id] = pet

	// Now, we have to return the NewPet
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(pet)
}

func (p *PetStore) FindPetByID(w http.ResponseWriter, r *http.Request, id int64) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[id]
	if !found {
		sendPetStoreError(w, http.StatusNotFound, fmt.Sprintf("Could not find pet with ID %d", id))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(pet)
}

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
