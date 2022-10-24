//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=types.cfg.yaml ../../petstore-expanded.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=server.cfg.yaml ../../petstore-expanded.yaml

package api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2"
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

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendPetStoreError(c *fiber.Ctx, code int, message string) error {

	petErr := Error{
		Code:    int32(code),
		Message: message,
	}

	return c.Status(code).JSON(petErr)
}

// FindPets implements all the handlers in the ServerInterface
func (p *PetStore) FindPets(c *fiber.Ctx, params FindPetsParams) error {

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

	return c.Status(http.StatusOK).JSON(result)
}

func (p *PetStore) AddPet(c *fiber.Ctx) error {

	// We expect a NewPet object in the request body.
	var newPet NewPet

	if err := c.BodyParser(&newPet); err != nil {
		return sendPetStoreError(c, http.StatusBadRequest, "Invalid format for NewPet")
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
	p.NextId = p.NextId + 1

	// Insert into map
	p.Pets[pet.Id] = pet

	// Now, we have to return the NewPet
	return c.Status(http.StatusCreated).JSON(pet)
}

func (p *PetStore) FindPetByID(c *fiber.Ctx, id int64) error {

	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[id]
	if !found {
		return sendPetStoreError(c, http.StatusNotFound, fmt.Sprintf("Could not find pet with ID %d", id))
	}

	return c.Status(http.StatusOK).JSON(pet)
}

func (p *PetStore) DeletePet(c *fiber.Ctx, id int64) error {

	p.Lock.Lock()
	defer p.Lock.Unlock()

	_, found := p.Pets[id]
	if !found {
		return sendPetStoreError(c, http.StatusNotFound, fmt.Sprintf("Could not find pet with ID %d", id))
	}
	delete(p.Pets, id)

	c.Status(http.StatusNoContent)
	return nil
}
