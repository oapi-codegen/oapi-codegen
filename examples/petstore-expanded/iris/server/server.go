package server

import (
	"errors"
	"net/http"

	"github.com/kataras/iris/v12"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/iris/api"
)

type PetStore struct {
	Store *Store
}

var _ api.ServerInterface = (*PetStore)(nil)

func NewPetStore() *PetStore {
	return &PetStore{
		Store: NewStore(),
	}
}

func sendPetStoreError(c iris.Context, code int, message string) {
	petErr := api.Error{
		Code:    int32(code),
		Message: message,
	}
	_ = c.StopWithJSON(code, petErr)
}

func (p *PetStore) FindPets(c iris.Context, params api.FindPetsParams) {
	result := p.Store.FindPets(params.Tags, params.Limit)
	_ = c.StopWithJSON(http.StatusOK, result)
}

func (p *PetStore) AddPet(c iris.Context) {
	var newPet api.NewPet
	if err := c.ReadJSON(&newPet); err != nil {
		sendPetStoreError(c, http.StatusBadRequest, "Invalid format for NewPet")
		return
	}

	pet := p.Store.AddPet(newPet)
	_ = c.StopWithJSON(http.StatusOK, pet)
}

func (p *PetStore) FindPetByID(c iris.Context, petId int64) {
	pet, err := p.Store.FindPetByID(petId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			sendPetStoreError(c, http.StatusNotFound, err.Error())
			return
		}
		sendPetStoreError(c, http.StatusInternalServerError, err.Error())
		return
	}
	_ = c.StopWithJSON(http.StatusOK, pet)
}

func (p *PetStore) DeletePet(c iris.Context, id int64) {
	err := p.Store.DeletePet(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			sendPetStoreError(c, http.StatusNotFound, err.Error())
			return
		}
		sendPetStoreError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.StatusCode(http.StatusNoContent)
}
