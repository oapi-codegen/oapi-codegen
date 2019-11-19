// Package api provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package api

// Error defines model for Error.
type Error struct {
	Code    int32  `json:"code" xml:"code"`
	Message string `json:"message" xml:"message"`
}

// NewPet defines model for NewPet.
type NewPet struct {
	Name string  `json:"name" xml:"name"`
	Tag  *string `json:"tag,omitempty" xml:"tag,omitempty"`
}

// Pet defines model for Pet.
type Pet struct {
	// Embedded struct due to allOf(#/components/schemas/NewPet)
	NewPet
	// Embedded fields due to inline allOf schema
	Id int64 `json:"id" xml:"id"`
}

// FindPetsParams defines parameters for FindPets.
type FindPetsParams struct {

	// tags to filter by
	Tags *[]string `json:"tags,omitempty" xml:"tags-list>tags,omitempty"`

	// maximum number of results to return
	Limit *int32 `json:"limit,omitempty" xml:"limit,omitempty"`
}

// addPetJSONBody defines parameters for AddPet.
type addPetJSONBody NewPet

// AddPetRequestBody defines body for AddPet for application/json ContentType.
type AddPetJSONRequestBody addPetJSONBody
