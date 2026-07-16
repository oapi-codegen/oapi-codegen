package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/strict/api"
)

type PetStore struct {
	Store *Store
}

var _ api.StrictServerInterface = (*PetStore)(nil)

func NewPetStore() *PetStore {
	return &PetStore{
		Store: NewStore(),
	}
}

func (p *PetStore) FindPets(ctx context.Context, request api.FindPetsRequestObject) (api.FindPetsResponseObject, error) {
	result := p.Store.FindPets(request.Params.Tags, request.Params.Limit)
	return api.FindPets200JSONResponse(result), nil
}

func (p *PetStore) AddPet(ctx context.Context, request api.AddPetRequestObject) (api.AddPetResponseObject, error) {
	pet := p.Store.AddPet(*request.Body)
	return api.AddPet200JSONResponse(pet), nil
}

func (p *PetStore) FindPetByID(ctx context.Context, request api.FindPetByIDRequestObject) (api.FindPetByIDResponseObject, error) {
	pet, err := p.Store.FindPetByID(request.Id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.FindPetByIDdefaultJSONResponse{
				StatusCode: http.StatusNotFound,
				Body:       api.Error{Code: http.StatusNotFound, Message: fmt.Sprintf("Could not find pet with ID %d", request.Id)},
			}, nil
		}
		return nil, err
	}
	return api.FindPetByID200JSONResponse(pet), nil
}

func (p *PetStore) DeletePet(ctx context.Context, request api.DeletePetRequestObject) (api.DeletePetResponseObject, error) {
	err := p.Store.DeletePet(request.Id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.DeletePetdefaultJSONResponse{
				StatusCode: http.StatusNotFound,
				Body:       api.Error{Code: http.StatusNotFound, Message: fmt.Sprintf("Could not find pet with ID %d", request.Id)},
			}, nil
		}
		return nil, err
	}
	return api.DeletePet204Response{}, nil
}
