package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/deepmap/oapi-codegen/examples/petstore-expanded/fiber/api"
)

func doGet(t *testing.T, app *fiber.App, rawURL string) (*http.Response, error) {

	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	req := httptest.NewRequest("GET", u.RequestURI(), nil)
	req.Header.Add("Accept", "application/json")
	req.Host = u.Host

	return app.Test(req)
}

func doPost(t *testing.T, app *fiber.App, rawURL string, jsonBody interface{}) (*http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	buf, err := json.Marshal(jsonBody)
	if err != nil {
		return nil, err
	}
	req := httptest.NewRequest("POST", u.RequestURI(), bytes.NewReader(buf))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Host = u.Host
	return app.Test(req)
}

func doDelete(t *testing.T, app *fiber.App, rawURL string) (*http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	req := httptest.NewRequest("DELETE", u.RequestURI(), nil)
	req.Header.Add("Accept", "application/json")
	req.Host = u.Host
	return app.Test(req)
}

func TestPetStore(t *testing.T) {
	var err error
	store := api.NewPetStore()
	fiberPetServer := NewFiberPetServer(store)

	t.Run("Add pet", func(t *testing.T) {
		tag := "TagOfSpot"
		newPet := api.NewPet{
			Name: "Spot",
			Tag:  &tag,
		}

		rr, _ := doPost(t, fiberPetServer, "/pets", newPet)
		assert.Equal(t, http.StatusCreated, rr.StatusCode)

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
		rr, _ := doGet(t, fiberPetServer, fmt.Sprintf("/pets/%d", pet.Id))
		assert.Equal(t, http.StatusOK, rr.StatusCode)

		var resultPet api.Pet
		err = json.NewDecoder(rr.Body).Decode(&resultPet)
		assert.NoError(t, err, "error getting pet")
		assert.Equal(t, pet, resultPet)
	})

	t.Run("Pet not found", func(t *testing.T) {
		rr, _ := doGet(t, fiberPetServer, "/pets/27179095781")
		assert.Equal(t, http.StatusNotFound, rr.StatusCode)

		var petError api.Error
		err = json.NewDecoder(rr.Body).Decode(&petError)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, int32(http.StatusNotFound), petError.Code)
	})

	t.Run("List all pets", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{1: {}, 2: {}}

		// Now, list all pets, we should have two
		rr, _ := doGet(t, fiberPetServer, "/pets")
		assert.Equal(t, http.StatusOK, rr.StatusCode)

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
		rr, _ := doGet(t, fiberPetServer, "/pets?tags=TagOfFido")
		assert.Equal(t, http.StatusOK, rr.StatusCode)

		var petList []api.Pet
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 1, len(petList))
	})

	t.Run("Filter pets by tag", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{1: {}, 2: {}}

		// Filter pets by non existent tag, we should have 0
		rr, _ := doGet(t, fiberPetServer, "/pets?tags=NotExists")
		assert.Equal(t, http.StatusOK, rr.StatusCode)

		var petList []api.Pet
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 0, len(petList))
	})

	t.Run("Delete pets", func(t *testing.T) {
		store.Pets = map[int64]api.Pet{1: {}, 2: {}}

		// Let's delete non-existent pet
		rr, _ := doDelete(t, fiberPetServer, "/pets/7")
		assert.Equal(t, http.StatusNotFound, rr.StatusCode)

		var petError api.Error
		err = json.NewDecoder(rr.Body).Decode(&petError)
		assert.NoError(t, err, "error unmarshaling PetError")
		assert.Equal(t, int32(http.StatusNotFound), petError.Code)

		// Now, delete both real pets
		rr, _ = doDelete(t, fiberPetServer, "/pets/1")
		assert.Equal(t, http.StatusNoContent, rr.StatusCode)

		rr, _ = doDelete(t, fiberPetServer, "/pets/2")
		assert.Equal(t, http.StatusNoContent, rr.StatusCode)

		// Should have no pets left.
		var petList []api.Pet
		rr, _ = doGet(t, fiberPetServer, "/pets")
		assert.Equal(t, http.StatusOK, rr.StatusCode)
		err = json.NewDecoder(rr.Body).Decode(&petList)
		assert.NoError(t, err, "error getting response", err)
		assert.Equal(t, 0, len(petList))
	})
}
