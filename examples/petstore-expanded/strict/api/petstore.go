//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=types.cfg.yaml ../../petstore-expanded.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=server.cfg.yaml ../../petstore-expanded.yaml

package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

type PetStore struct {
	Pets   map[int64]Pet
	NextId int64
	Lock   sync.Mutex
}

// Make sure we conform to StrictServerInterface

var _ StrictServerInterface = (*PetStore)(nil)

func NewPetStore() *PetStore {
	return &PetStore{
		Pets:   make(map[int64]Pet),
		NextId: 1000,
	}
}

// Here, we implement all of the handlers in the ServerInterface
func (p *PetStore) FindPets(ctx context.Context, request FindPetsRequestObject) (FindPetsResponseObject, error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var result []Pet

	for _, pet := range p.Pets {
		if request.Params.Tags != nil {
			// If we have tags,  filter pets by tag
			for _, t := range *request.Params.Tags {
				if pet.Tag != nil && (*pet.Tag == t) {
					result = append(result, pet)
				}
			}
		} else {
			// Add all pets if we're not filtering
			result = append(result, pet)
		}

		if request.Params.Limit != nil {
			l := int(*request.Params.Limit)
			if len(result) >= l {
				// We're at the limit
				break
			}
		}
	}

	return FindPets200JSONResponse(result), nil
}

func (p *PetStore) AddPet(ctx context.Context, request AddPetRequestObject) (AddPetResponseObject, error) {
	// We now have a pet, let's add it to our "database".
	// We're always asynchronous, so lock unsafe operations below
	p.Lock.Lock()
	defer p.Lock.Unlock()

	// We handle pets, not NewPets, which have an additional ID field
	var pet Pet
	pet.Name = request.Body.Name
	pet.Tag = request.Body.Tag
	pet.Id = p.NextId
	p.NextId = p.NextId + 1

	// Insert into map
	p.Pets[pet.Id] = pet

	// Now, we have to return the NewPet
	return AddPet200JSONResponse(pet), nil
}

func (p *PetStore) FindPetByID(ctx context.Context, request FindPetByIDRequestObject) (FindPetByIDResponseObject, error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[request.Id]
	if !found {
		return FindPetByIDdefaultJSONResponse{StatusCode: http.StatusNotFound, Body: Error{Code: http.StatusNotFound, Message: fmt.Sprintf("Could not find pet with ID %d", request.Id)}}, nil
	}

	return FindPetByID200JSONResponse(pet), nil
}

func (p *PetStore) DeletePet(ctx context.Context, request DeletePetRequestObject) (DeletePetResponseObject, error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	_, found := p.Pets[request.Id]
	if !found {
		return DeletePetdefaultJSONResponse{StatusCode: http.StatusNotFound, Body: Error{Code: http.StatusNotFound, Message: fmt.Sprintf("Could not find pet with ID %d", request.Id)}}, nil
	}
	delete(p.Pets, request.Id)

	return DeletePet204Response{}, nil
}
