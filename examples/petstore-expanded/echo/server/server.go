package server

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/models"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/store"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/echo/api"
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

func sendPetStoreError(ctx echo.Context, code int, message string) error {
	petErr := models.Error{
		Code:    int32(code),
		Message: message,
	}
	return ctx.JSON(code, petErr)
}

func (p *PetStore) FindPets(ctx echo.Context, params models.FindPetsParams) error {
	result := p.Store.FindPets(params.Tags, params.Limit)
	return ctx.JSON(http.StatusOK, result)
}

func (p *PetStore) AddPet(ctx echo.Context) error {
	var newPet models.NewPet
	if err := ctx.Bind(&newPet); err != nil {
		return sendPetStoreError(ctx, http.StatusBadRequest, "Invalid format for NewPet")
	}

	pet := p.Store.AddPet(newPet)
	return ctx.JSON(http.StatusOK, pet)
}

func (p *PetStore) FindPetByID(ctx echo.Context, petId int64) error {
	pet, err := p.Store.FindPetByID(petId)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return sendPetStoreError(ctx, http.StatusNotFound, err.Error())
		}
		return sendPetStoreError(ctx, http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, pet)
}

func (p *PetStore) DeletePet(ctx echo.Context, id int64) error {
	err := p.Store.DeletePet(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return sendPetStoreError(ctx, http.StatusNotFound, err.Error())
		}
		return sendPetStoreError(ctx, http.StatusInternalServerError, err.Error())
	}
	return ctx.NoContent(http.StatusNoContent)
}
