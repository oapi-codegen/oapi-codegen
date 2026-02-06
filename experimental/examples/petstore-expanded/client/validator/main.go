package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	petstore "github.com/oapi-codegen/oapi-codegen/experimental/examples/petstore-expanded"
	"github.com/oapi-codegen/oapi-codegen/experimental/examples/petstore-expanded/client"
)

func ptr[T any](v T) *T { return &v }

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Petstore server URL")
	flag.Parse()

	log.SetFlags(0)
	log.Printf("Petstore Validator")
	log.Printf("==================")
	log.Printf("Server: %s", *serverURL)
	log.Println()

	c, err := client.NewSimpleClient(*serverURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	// Step 1: Add Fido the Dog and Sushi the Cat
	log.Println("--- Step 1: Creating pets ---")

	fido, err := c.AddPet(ctx, petstore.NewPet{Name: "Fido", Tag: ptr("Dog")})
	if err != nil {
		log.Fatalf("Failed to create Fido: %v", err)
	}
	log.Printf("Created pet: %s (tag=%s, id=%d)", fido.Name, derefTag(fido.Tag), fido.ID)

	sushi, err := c.AddPet(ctx, petstore.NewPet{Name: "Sushi", Tag: ptr("Cat")})
	if err != nil {
		log.Fatalf("Failed to create Sushi: %v", err)
	}
	log.Printf("Created pet: %s (tag=%s, id=%d)", sushi.Name, derefTag(sushi.Tag), sushi.ID)
	log.Println()

	// Step 2: List all pets
	log.Println("--- Step 2: Listing all pets ---")
	pets, err := c.FindPets(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list pets: %v", err)
	}
	printPets(pets)
	log.Println()

	// Step 3: Delete Fido
	log.Printf("--- Step 3: Deleting Fido (id=%d) ---", fido.ID)
	resp, err := c.Client.DeletePet(ctx, fido.ID)
	if err != nil {
		log.Fatalf("Failed to delete Fido: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode == 204 {
		log.Printf("Deleted Fido successfully (HTTP %d)", resp.StatusCode)
	} else {
		log.Fatalf("Unexpected status deleting Fido: HTTP %d", resp.StatusCode)
	}
	log.Println()

	// Step 4: Add Slimy the Lizard
	log.Println("--- Step 4: Creating Slimy the Lizard ---")
	slimy, err := c.AddPet(ctx, petstore.NewPet{Name: "Slimy", Tag: ptr("Lizard")})
	if err != nil {
		log.Fatalf("Failed to create Slimy: %v", err)
	}
	log.Printf("Created pet: %s (tag=%s, id=%d)", slimy.Name, derefTag(slimy.Tag), slimy.ID)
	log.Println()

	// Step 5: List all pets again
	log.Println("--- Step 5: Listing all pets (after changes) ---")
	pets, err = c.FindPets(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list pets: %v", err)
	}
	printPets(pets)
	log.Println()

	// Validate final state
	log.Println("--- Validation ---")
	ok := true
	if len(pets) != 2 {
		log.Printf("FAIL: expected 2 pets, got %d", len(pets))
		ok = false
	}
	names := map[string]bool{}
	for _, p := range pets {
		names[p.Name] = true
	}
	if !names["Sushi"] {
		log.Printf("FAIL: Sushi not found in pet list")
		ok = false
	}
	if !names["Slimy"] {
		log.Printf("FAIL: Slimy not found in pet list")
		ok = false
	}
	if names["Fido"] {
		log.Printf("FAIL: Fido should have been deleted but is still present")
		ok = false
	}

	if ok {
		log.Println("PASS: All validations passed!")
	} else {
		log.Println("FAIL: Some validations failed")
		os.Exit(1)
	}
}

func derefTag(tag *string) string {
	if tag == nil {
		return "<none>"
	}
	return *tag
}

func printPets(pets []petstore.Pet) {
	if len(pets) == 0 {
		log.Println("  (no pets)")
		return
	}
	maxName := 0
	for _, p := range pets {
		if len(p.Name) > maxName {
			maxName = len(p.Name)
		}
	}
	for _, p := range pets {
		padding := strings.Repeat(" ", maxName-len(p.Name))
		log.Printf("  - %s%s  tag=%-8s  id=%d", p.Name, padding, derefTag(p.Tag), p.ID)
	}
}
