package server

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/models"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/store"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/fiber/api"
)

type PetStore struct {
	Store *store.PetStore
}

var _ api.ServerInterface = (*PetStore)(nil)

func NewPetStore() *PetStore {
	return &PetStore{
		Store: store.NewPetStore(),
	}
}

func sendPetStoreError(c *fiber.Ctx, code int, message string) error {
	petErr := models.Error{
		Code:    int32(code),
		Message: message,
	}
	return c.Status(code).JSON(petErr)
}

func (p *PetStore) FindPets(c *fiber.Ctx, params models.FindPetsParams) error {
	result := p.Store.FindPets(params.Tags, params.Limit)
	return c.Status(http.StatusOK).JSON(result)
}

func (p *PetStore) AddPet(c *fiber.Ctx) error {
	var newPet models.NewPet
	if err := c.BodyParser(&newPet); err != nil {
		return sendPetStoreError(c, http.StatusBadRequest, "Invalid format for NewPet")
	}

	pet := p.Store.AddPet(newPet)
	return c.Status(http.StatusOK).JSON(pet)
}

func (p *PetStore) FindPetByID(c *fiber.Ctx, id int64) error {
	pet, err := p.Store.FindPetByID(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return sendPetStoreError(c, http.StatusNotFound, err.Error())
		}
		return sendPetStoreError(c, http.StatusInternalServerError, err.Error())
	}
	return c.Status(http.StatusOK).JSON(pet)
}

func (p *PetStore) DeletePet(c *fiber.Ctx, id int64) error {
	err := p.Store.DeletePet(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return sendPetStoreError(c, http.StatusNotFound, err.Error())
		}
		return sendPetStoreError(c, http.StatusInternalServerError, err.Error())
	}
	c.Status(http.StatusNoContent)
	return nil
}
