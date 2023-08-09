package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deepmap/oapi-codegen/examples/petstore-expanded/strict/api"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doGet(t *testing.T, mux *chi.Mux, url string) *httptest.ResponseRecorder {
	response := testutil.NewRequest().Get(url).WithAcceptJson().GoWithHTTPHandler(t, mux)
	return response.Recorder
}

func TestPetStore(t *testing.T) {
	var err error

	// Get the swagger description of our API
	swagger, err := api.GetSwagger()
	require.NoError(t, err)

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// This is how you set up a basic chi router
	r := chi.NewRouter()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(middleware.OapiRequestValidator(swagger))

	store := api.NewPetStore()
	api.HandlerFromMux(api.NewStrictHandler(store, nil), r)

	t.Run("Add pet", func(t *testing.T) {
		tag := "TagOfSpot"
		newPet := api.NewPet{
			Name: "Spot",
			Tag:  &tag,
		}

		rr := testutil.NewRequest().Post("/pets").WithJsonBody(newPet).GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)

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
		rr := doGet(t, r, fmt.Sprintf("/pets/%d", pet.Id))

		var resultPet api.Pet
		err = json.NewDecoder(rr.Body).Decode(&resultPet)
		assert.NoError(t, err, "error getting pet")
		assert.Equal(t, pet, resultPet)
	})

	t.Run("Pet not found", func(t *testing.T) {
		rr := doGet(t, r, "/pets/27179095781")
		assert.Equal(t, http.StatusNotFound, rr.Code)

		var petError api.Error
		err = json.NewDecoder(rr.Body).Decode(&petError)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, int32(http.StatusNotFound), petError.Code)
	})

	t.Run("List all pets", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{
			1: {},
			2: {},
		}

		// Now, list all pets, we should have two
		rr := doGet(t, r, "/pets")
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err = json.NewDecoder(rr.Body).Decode(&petList)
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
		rr := doGet(t, r, "/pets?tags=TagOfFido")
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 1, len(petList))
	})

	t.Run("Filter pets by tag", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{
			1: {},
			2: {},
		}

		// Filter pets by non-existent tag, we should have 0
		rr := doGet(t, r, "/pets?tags=NotExists")
		assert.Equal(t, http.StatusOK, rr.Code)

		var petList []api.Pet
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 0, len(petList))
	})

	t.Run("Delete pets", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{
			1: {},
			2: {},
		}

		// Let's delete non-existent pet
		rr := testutil.NewRequest().Delete("/pets/7").GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusNotFound, rr.Code)

		var petError api.Error
		err = json.NewDecoder(rr.Body).Decode(&petError)
		assert.NoError(t, err, "error unmarshaling PetError")
		assert.Equal(t, int32(http.StatusNotFound), petError.Code)

		// Now, delete both real pets
		rr = testutil.NewRequest().Delete("/pets/1").GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusNoContent, rr.Code)

		rr = testutil.NewRequest().Delete("/pets/2").GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Should have no pets left.
		var petList []api.Pet
		rr = doGet(t, r, "/pets")
		assert.Equal(t, http.StatusOK, rr.Code)
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 0, len(petList))
	})
}
