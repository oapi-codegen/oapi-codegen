//go:build go1.22

package server

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v5"
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
func sendPetStoreError(ctx *echo.Context, code int, message string) error {
	petErr := petstore.Error{
		Code:    int32(code),
		Message: message,
	}
	return ctx.JSON(code, petErr)
}

// FindPets returns all pets, optionally filtered by tags and limited.
func (p *PetStore) FindPets(ctx *echo.Context, params FindPetsParams) error {
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

	return ctx.JSON(http.StatusOK, result)
}

// AddPet creates a new pet.
func (p *PetStore) AddPet(ctx *echo.Context) error {
	// We expect a NewPet object in the request body.
	var newPet petstore.NewPet
	if err := ctx.Bind(&newPet); err != nil {
		return sendPetStoreError(ctx, http.StatusBadRequest, "Invalid format for NewPet")
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
	return ctx.JSON(http.StatusCreated, pet)
}

// FindPetByID returns a pet by ID.
func (p *PetStore) FindPetByID(ctx *echo.Context, id int64) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[id]
	if !found {
		return sendPetStoreError(ctx, http.StatusNotFound, "Could not find pet with ID")
	}

	return ctx.JSON(http.StatusOK, pet)
}

// DeletePet deletes a pet by ID.
func (p *PetStore) DeletePet(ctx *echo.Context, id int64) error {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	_, found := p.Pets[id]
	if !found {
		return sendPetStoreError(ctx, http.StatusNotFound, "Could not find pet with ID")
	}
	delete(p.Pets, id)

	return ctx.NoContent(http.StatusNoContent)
}
