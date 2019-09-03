package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deepmap/oapi-codegen/examples/petstore-expanded/chi/api"
)

func DoJson(handler http.Handler, method string, url string, body interface{}) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(b))
	handler.ServeHTTP(rr, req)

	return rr
}

func TestPetStore(t *testing.T) {
	var err error

	store := api.NewPetStore()
	h := api.Handler(store)

	t.Run("Add pet", func(t *testing.T) {
		tag := "TagOfSpot"
		newPet := api.NewPet{
			Name: "Spot",
			Tag:  &tag,
		}

		rr := DoJson(h, "POST", "/pets", newPet)
		assert.Equal(t, http.StatusCreated, rr.Code)

		var resultPet api.Pet
		err = json.NewDecoder(rr.Body).Decode(&resultPet)
		assert.NoError(t, err, "error unmarshaling response")
		assert.Equal(t, newPet.Name, resultPet.Name)
		assert.Equal(t, *newPet.Tag, *resultPet.Tag)
	})

	t.Run("Find pet by ID", func(t *testing.T) {
		pet := api.Pet{
			Id: 100,
		}

		store.Pets[pet.Id] = pet
		rr := DoJson(h, "GET", fmt.Sprintf("/pets/%d", pet.Id), nil)

		var resultPet api.Pet
		err = json.NewDecoder(rr.Body).Decode(&resultPet)
		assert.NoError(t, err, "error getting pet")
		assert.Equal(t, pet, resultPet)
	})

	t.Run("Pet not found", func(t *testing.T) {
		rr := DoJson(h, "GET", "/pets/27179095781", nil)
		assert.Equal(t, http.StatusNotFound, rr.Code)

		var petError api.Error
		err = json.NewDecoder(rr.Body).Decode(&petError)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, int32(http.StatusNotFound), petError.Code)
	})

	t.Run("List all pets", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{
			1: api.Pet{},
			2: api.Pet{},
		}

		// Now, list all pets, we should have two
		rr := DoJson(h, "GET", "/pets", nil)
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 2, len(petList))
	})

	t.Run("Filter pets by tag", func(t *testing.T) {
		tag := "TagOfFido"

		store.Pets = map[int64]api.Pet{
			1: api.Pet{
				NewPet: api.NewPet{
					Tag: &tag,
				},
			},
			2: api.Pet{},
		}

		// Filter pets by tag, we should have 1
		rr := DoJson(h, "GET", "/pets?tags=TagOfFido", nil)
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 1, len(petList))
	})

	t.Run("Filter pets by tag", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{
			1: api.Pet{},
			2: api.Pet{},
		}

		// Filter pets by non existent tag, we should have 0
		rr := DoJson(h, "GET", "/pets?tags=NotExists", nil)
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 0, len(petList))
	})

	t.Run("Delete pets", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{
			1: api.Pet{},
			2: api.Pet{},
		}

		// Let's delete non-existent pet
		rr := DoJson(h, "DELETE", "/pets/7", nil)
		assert.Equal(t, http.StatusNotFound, rr.Code)

		var petError api.Error
		err = json.NewDecoder(rr.Body).Decode(&petError)
		assert.NoError(t, err, "error unmarshaling PetError")
		assert.Equal(t, int32(http.StatusNotFound), petError.Code)

		// Now, delete both real pets
		rr = DoJson(h, "DELETE", "/pets/1", nil)
		assert.Equal(t, http.StatusNoContent, rr.Code)

		rr = DoJson(h, "DELETE", "/pets/2", nil)
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Should have no pets left.
		var petList []api.Pet
		rr = DoJson(h, "GET", "/pets", nil)
		assert.Equal(t, http.StatusOK, rr.Code)
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 0, len(petList))
	})
}
