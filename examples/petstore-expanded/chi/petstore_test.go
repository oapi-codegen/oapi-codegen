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

	h := api.Handler(api.NewPetStore())

	// At this point, we can start sending simulated Http requests, and record
	// the HTTP responses to check for validity. This exercises every part of
	// the stack except the well-tested HTTP system in Go, which there is no
	// point for us to test.

	tag := "TagOfSpot"
	newPet := api.NewPet{
		Name: "Spot",
		Tag:  &tag,
	}

	rr := DoJson(h, "POST", "/pets", newPet)

	// We expect 201 code on successful pet insertion
	assert.Equal(t, http.StatusCreated, rr.Code)

	// We should have gotten a response from the server with the new pet. Make
	// sure that its fields match.
	var resultPet api.Pet

	err = json.NewDecoder(rr.Body).Decode(&resultPet)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, newPet.Name, resultPet.Name)
	assert.Equal(t, *newPet.Tag, *resultPet.Tag)

	// This is the Id of the pet we inserted.
	petId := resultPet.Id

	// Test the getter function.
	rr = DoJson(h, "GET", fmt.Sprintf("/pets/%d", petId), nil)

	var resultPet2 api.Pet
	err = json.NewDecoder(rr.Body).Decode(&resultPet2)
	assert.NoError(t, err, "error getting pet")
	assert.Equal(t, resultPet, resultPet2)

	// We should get a 404 on invalid ID
	rr = DoJson(h, "GET", "/pets/27179095781", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	var petError api.Error
	err = json.NewDecoder(rr.Body).Decode(&petError)
	assert.NoError(t, err, "error getting response", err)
	assert.Equal(t, int32(http.StatusNotFound), petError.Code)

	// Let's insert another pet for subsequent tests.
	tag = "TagOfFido"
	newPet = api.NewPet{
		Name: "Fido",
		Tag:  &tag,
	}
	rr = DoJson(h, "POST", "/pets", newPet)
	// We expect 201 code on successful pet insertion
	assert.Equal(t, http.StatusCreated, rr.Code)
	// We should have gotten a response from the server with the new pet. Make
	// sure that its fields match.
	err = json.NewDecoder(rr.Body).Decode(&resultPet)
	assert.NoError(t, err, "error unmarshaling response")
	petId2 := resultPet.Id

	// Now, list all pets, we should have two
	rr = DoJson(h, "GET", "/pets", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	var petList []api.Pet
	err = json.NewDecoder(rr.Body).Decode(&petList)
	assert.NoError(t, err, "error getting response", err)
	assert.Equal(t, 2, len(petList))

	// Filter pets by tag, we should have 1
	petList = nil
	rr = DoJson(h, "GET", "/pets?tags=TagOfFido", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	err = json.NewDecoder(rr.Body).Decode(&petList)
	assert.NoError(t, err, "error getting response", err)
	assert.Equal(t, 1, len(petList))

	// Filter pets by non existent tag, we should have 0
	petList = nil
	rr = DoJson(h, "GET", "/pets?tags=NotExists", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	err = json.NewDecoder(rr.Body).Decode(&petList)
	assert.NoError(t, err, "error getting response", err)
	assert.Equal(t, 0, len(petList))

	// Let's delete non-existent pet
	rr = DoJson(h, "DELETE", "/pets/7", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	err = json.NewDecoder(rr.Body).Decode(&petError)
	assert.NoError(t, err, "error unmarshaling PetError")
	assert.Equal(t, int32(http.StatusNotFound), petError.Code)

	// Now, delete both real pets
	rr = DoJson(h, "DELETE", fmt.Sprintf("/pets/%d", petId), nil)
	assert.Equal(t, http.StatusNoContent, rr.Code)
	rr = DoJson(h, "DELETE", fmt.Sprintf("/pets/%d", petId2), nil)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Should have no pets left.
	petList = nil
	rr = DoJson(h, "GET", "/pets", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	err = json.NewDecoder(rr.Body).Decode(&petList)
	assert.NoError(t, err, "error getting response", err)
	assert.Equal(t, 0, len(petList))
}
