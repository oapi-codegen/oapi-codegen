package output

import (
	"encoding/json"
	"testing"
)

// TestAnyOfInlineCatType verifies the Cat type fields are accessible.
// V2 test suite: internal/test/components/anyof/inline
func TestAnyOfInlineCatType(t *testing.T) {
	id := "cat-1"
	name := "Whiskers"
	breed := "Siamese"
	color := "cream"
	purrs := true

	cat := Cat{
		ID:    &id,
		Name:  &name,
		Breed: &breed,
		Color: &color,
		Purrs: &purrs,
	}

	if *cat.ID != "cat-1" {
		t.Errorf("ID = %q, want %q", *cat.ID, "cat-1")
	}
	if *cat.Name != "Whiskers" {
		t.Errorf("Name = %q, want %q", *cat.Name, "Whiskers")
	}
	if *cat.Purrs != true {
		t.Errorf("Purrs = %v, want true", *cat.Purrs)
	}
}

// TestAnyOfInlineDogType verifies the Dog type fields are accessible.
func TestAnyOfInlineDogType(t *testing.T) {
	id := "dog-1"
	name := "Rex"
	barks := true

	dog := Dog{
		ID:    &id,
		Name:  &name,
		Barks: &barks,
	}

	if *dog.ID != "dog-1" {
		t.Errorf("ID = %q, want %q", *dog.ID, "dog-1")
	}
	if *dog.Barks != true {
		t.Errorf("Barks = %v, want true", *dog.Barks)
	}
}

// TestAnyOfInlineRatType verifies the Rat type fields are accessible.
func TestAnyOfInlineRatType(t *testing.T) {
	id := "rat-1"
	name := "Remy"
	squeaks := true

	rat := Rat{
		ID:      &id,
		Name:    &name,
		Squeaks: &squeaks,
	}

	if *rat.ID != "rat-1" {
		t.Errorf("ID = %q, want %q", *rat.ID, "rat-1")
	}
	if *rat.Squeaks != true {
		t.Errorf("Squeaks = %v, want true", *rat.Squeaks)
	}
}

// TestAnyOfInlineUnionType verifies the anyOf union type
// GetPets200ResponseJSON2 holds Cat, Dog, and Rat members.
func TestAnyOfInlineUnionType(t *testing.T) {
	id := "cat-1"
	name := "Whiskers"
	cat := Cat{
		ID:   &id,
		Name: &name,
	}

	union := GetPets200ResponseJSON2{
		Cat: &cat,
	}

	if union.Cat == nil {
		t.Fatal("Cat should not be nil")
	}
	if *union.Cat.ID != "cat-1" {
		t.Errorf("Cat.ID = %q, want %q", *union.Cat.ID, "cat-1")
	}
	if union.Dog != nil {
		t.Error("Dog should be nil")
	}
	if union.Rat != nil {
		t.Error("Rat should be nil")
	}
}

// TestAnyOfInlineUnionMarshalJSON verifies that MarshalJSON merges the fields
// from the set anyOf member into a single JSON object.
func TestAnyOfInlineUnionMarshalJSON(t *testing.T) {
	id := "dog-1"
	name := "Buddy"
	barks := true
	dog := Dog{
		ID:    &id,
		Name:  &name,
		Barks: &barks,
	}

	union := GetPets200ResponseJSON2{
		Dog: &dog,
	}

	data, err := json.Marshal(union)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal into map failed: %v", err)
	}

	if m["id"] != "dog-1" {
		t.Errorf("id = %v, want %q", m["id"], "dog-1")
	}
	if m["name"] != "Buddy" {
		t.Errorf("name = %v, want %q", m["name"], "Buddy")
	}
	if m["barks"] != true {
		t.Errorf("barks = %v, want true", m["barks"])
	}
}

// TestAnyOfInlineUnionUnmarshalJSON verifies that UnmarshalJSON populates all
// matching anyOf members from the input JSON.
func TestAnyOfInlineUnionUnmarshalJSON(t *testing.T) {
	input := `{"id":"pet-1","name":"Furball","color":"brown","purrs":true}`

	var union GetPets200ResponseJSON2
	if err := json.Unmarshal([]byte(input), &union); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Cat should match because purrs is a Cat field
	if union.Cat == nil {
		t.Fatal("Cat should not be nil after unmarshal")
	}
	if *union.Cat.Name != "Furball" {
		t.Errorf("Cat.Name = %q, want %q", *union.Cat.Name, "Furball")
	}
	if *union.Cat.Purrs != true {
		t.Errorf("Cat.Purrs = %v, want true", *union.Cat.Purrs)
	}

	// Dog and Rat should also match (anyOf allows multiple matches)
	if union.Dog == nil {
		t.Fatal("Dog should not be nil (anyOf allows multiple matches)")
	}
	if union.Rat == nil {
		t.Fatal("Rat should not be nil (anyOf allows multiple matches)")
	}
}

// TestAnyOfInlineResponseType verifies the GetPetsJSONResponse wrapper type.
func TestAnyOfInlineResponseType(t *testing.T) {
	id := "rat-1"
	name := "Scabbers"
	rat := Rat{
		ID:   &id,
		Name: &name,
	}

	resp := GetPetsJSONResponse{
		Data: []GetPets200ResponseJSON2{
			{Rat: &rat},
		},
	}

	if len(resp.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(resp.Data))
	}
	if resp.Data[0].Rat == nil {
		t.Fatal("Data[0].Rat should not be nil")
	}
	if *resp.Data[0].Rat.Name != "Scabbers" {
		t.Errorf("Data[0].Rat.Name = %q, want %q", *resp.Data[0].Rat.Name, "Scabbers")
	}
}

// TestAnyOfInlineResponseJSONRoundTrip verifies JSON round-trip for the
// response wrapper containing anyOf union items.
func TestAnyOfInlineResponseJSONRoundTrip(t *testing.T) {
	id := "cat-2"
	name := "Luna"
	purrs := true
	cat := Cat{ID: &id, Name: &name, Purrs: &purrs}

	original := GetPetsJSONResponse{
		Data: []GetPets200ResponseJSON2{
			{Cat: &cat},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded GetPetsJSONResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(decoded.Data))
	}
	if decoded.Data[0].Cat == nil {
		t.Fatal("decoded Cat should not be nil")
	}
	if *decoded.Data[0].Cat.Name != "Luna" {
		t.Errorf("Cat.Name = %q, want %q", *decoded.Data[0].Cat.Name, "Luna")
	}
}

// TestAnyOfInlineTypeAlias verifies the type alias for the data array.
func TestAnyOfInlineTypeAlias(t *testing.T) {
	var items GetPets200ResponseJSON1
	items = append(items, GetPets200ResponseJSON2{})
	if len(items) != 1 {
		t.Errorf("items length = %d, want 1", len(items))
	}
}

// TestAnyOfInlineApplyDefaults verifies that ApplyDefaults can be called on
// all types without panic.
func TestAnyOfInlineApplyDefaults(t *testing.T) {
	cat := &Cat{}
	cat.ApplyDefaults()

	dog := &Dog{}
	dog.ApplyDefaults()

	rat := &Rat{}
	rat.ApplyDefaults()

	resp := &GetPetsJSONResponse{}
	resp.ApplyDefaults()

	id := "test"
	union := &GetPets200ResponseJSON2{
		Cat: &Cat{ID: &id},
	}
	union.ApplyDefaults()
}
