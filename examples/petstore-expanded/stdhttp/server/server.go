package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/models"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/store"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/stdhttp/api"
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

func sendPetStoreError(w http.ResponseWriter, code int, message string) {
	petErr := models.Error{
		Code:    int32(code),
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(petErr)
}

func (p *PetStore) FindPets(w http.ResponseWriter, r *http.Request, params models.FindPetsParams) {
	result := p.Store.FindPets(params.Tags, params.Limit)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func (p *PetStore) AddPet(w http.ResponseWriter, r *http.Request) {
	var newPet models.NewPet
	if err := json.NewDecoder(r.Body).Decode(&newPet); err != nil {
		sendPetStoreError(w, http.StatusBadRequest, "Invalid format for NewPet")
		return
	}

	pet := p.Store.AddPet(newPet)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(pet)
}

func (p *PetStore) FindPetByID(w http.ResponseWriter, r *http.Request, id int64) {
	pet, err := p.Store.FindPetByID(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			sendPetStoreError(w, http.StatusNotFound, err.Error())
			return
		}
		sendPetStoreError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(pet)
}

func (p *PetStore) DeletePet(w http.ResponseWriter, r *http.Request, id int64) {
	err := p.Store.DeletePet(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			sendPetStoreError(w, http.StatusNotFound, err.Error())
			return
		}
		sendPetStoreError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
