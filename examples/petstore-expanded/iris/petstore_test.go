// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deepmap/oapi-codegen/examples/petstore-expanded/iris/api"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func doGet(t *testing.T, handler http.Handler, url string) *httptest.ResponseRecorder {
	response := testutil.NewRequest().Get(url).WithAcceptJson().GoWithHTTPHandler(t, handler)
	return response.Recorder
}

func TestPetStore(t *testing.T) {
	store := api.NewPetStore()
	irisPetServer := NewIrisPetServer(store, 8080)

	t.Run("Add pet", func(t *testing.T) {
		tag := "TagOfSpot"
		newPet := api.NewPet{
			Name: "Spot",
			Tag:  &tag,
		}

		rr := testutil.NewRequest().Post("/pets").WithJsonBody(newPet).GoWithHTTPHandler(t, irisPetServer).Recorder
		assert.Equal(t, http.StatusCreated, rr.Code)

		var resultPet api.Pet
		err := json.NewDecoder(rr.Body).Decode(&resultPet)
		assert.NoError(t, err, "error unmarshaling response")
		assert.Equal(t, newPet.Name, resultPet.Name)
		assert.Equal(t, *newPet.Tag, *resultPet.Tag)
	})

	t.Run("Find pet by ID", func(t *testing.T) {
		pet := api.Pet{
			Id: 100,
		}
		store.Pets[pet.Id] = pet
		rr := doGet(t, irisPetServer, fmt.Sprintf("/pets/%d", pet.Id))

		var resultPet api.Pet
		err := json.NewDecoder(rr.Body).Decode(&resultPet)
		assert.NoError(t, err, "error getting pet")
		assert.Equal(t, pet, resultPet)
	})

	t.Run("Pet not found", func(t *testing.T) {
		rr := doGet(t, irisPetServer, "/pets/27179095781")
		assert.Equal(t, http.StatusNotFound, rr.Code)

		var petError api.Error
		err := json.NewDecoder(rr.Body).Decode(&petError)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, int32(http.StatusNotFound), petError.Code)
	})

	t.Run("List all pets", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{1: {}, 2: {}}

		// Now, list all pets, we should have two
		rr := doGet(t, irisPetServer, "/pets")
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err := json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 2, len(petList))
	})

	t.Run("Filter pets by tag", func(t *testing.T) {
		tag := "TagOfFido"

		store.Pets = map[int64]api.Pet{
			1: {
				Tag: &tag,
			},
			2: {},
		}

		// Filter pets by tag, we should have 1
		rr := doGet(t, irisPetServer, "/pets?tags=TagOfFido")
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err := json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 1, len(petList))
	})

	t.Run("Filter pets by tag", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{1: {}, 2: {}}

		// Filter pets by non existent tag, we should have 0
		rr := doGet(t, irisPetServer, "/pets?tags=NotExists")
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err := json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 0, len(petList))
	})

	t.Run("Delete pets", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{1: {}, 2: {}}

		// Let's delete non-existent pet
		rr := testutil.NewRequest().Delete("/pets/7").GoWithHTTPHandler(t, irisPetServer).Recorder
		assert.Equal(t, http.StatusNotFound, rr.Code)

		var petError api.Error
		err := json.NewDecoder(rr.Body).Decode(&petError)
		assert.NoError(t, err, "error unmarshaling PetError")
		assert.Equal(t, int32(http.StatusNotFound), petError.Code)

		// Now, delete both real pets
		rr = testutil.NewRequest().Delete("/pets/1").GoWithHTTPHandler(t, irisPetServer).Recorder
		assert.Equal(t, http.StatusNoContent, rr.Code)

		rr = testutil.NewRequest().Delete("/pets/2").GoWithHTTPHandler(t, irisPetServer).Recorder
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Should have no pets left.
		var petList []api.Pet
		rr = doGet(t, irisPetServer, "/pets")
		assert.Equal(t, http.StatusOK, rr.Code)
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 0, len(petList))
	})
}
