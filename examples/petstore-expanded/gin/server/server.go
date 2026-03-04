package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/models"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/store"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/gin/api"
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

func sendPetStoreError(c *gin.Context, code int, message string) {
	petErr := models.Error{
		Code:    int32(code),
		Message: message,
	}
	c.JSON(code, petErr)
}

func (p *PetStore) FindPets(c *gin.Context, params models.FindPetsParams) {
	result := p.Store.FindPets(params.Tags, params.Limit)
	c.JSON(http.StatusOK, result)
}

func (p *PetStore) AddPet(c *gin.Context) {
	var newPet models.NewPet
	if err := c.Bind(&newPet); err != nil {
		sendPetStoreError(c, http.StatusBadRequest, "Invalid format for NewPet")
		return
	}

	pet := p.Store.AddPet(newPet)
	c.JSON(http.StatusOK, pet)
}

func (p *PetStore) FindPetByID(c *gin.Context, petId int64) {
	pet, err := p.Store.FindPetByID(petId)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			sendPetStoreError(c, http.StatusNotFound, err.Error())
			return
		}
		sendPetStoreError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, pet)
}

func (p *PetStore) DeletePet(c *gin.Context, id int64) {
	err := p.Store.DeletePet(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			sendPetStoreError(c, http.StatusNotFound, err.Error())
			return
		}
		sendPetStoreError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
