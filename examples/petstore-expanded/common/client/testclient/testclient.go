// Package testclient exercises all CRUD operations of the petstore API
// against a running server. It is used by both the CLI test client and
// per-variant integration tests.
package testclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/client/openapi"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/models"
)

// Run executes the full petstore test sequence against serverURL.
// It returns nil if all checks pass, or an error describing the failures.
func Run(serverURL string) error {
	client, err := openapi.NewClientWithResponses(serverURL)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx := context.Background()
	failures := 0

	// 1. Add pet "Spot" with tag
	fmt.Println("1. Adding pet Spot...")
	tag := "TagOfSpot"
	addResp, err := client.AddPetWithResponse(ctx, models.NewPet{
		Name: "Spot",
		Tag:  &tag,
	})
	if err != nil {
		return fmt.Errorf("AddPet failed: %w", err)
	}
	if addResp.StatusCode() != http.StatusOK {
		failures++
		fmt.Printf("   FAIL: expected 200, got %d\n", addResp.StatusCode())
	} else if addResp.JSON200 == nil {
		failures++
		fmt.Println("   FAIL: response body is nil")
	} else {
		if addResp.JSON200.Name != "Spot" {
			failures++
			fmt.Printf("   FAIL: expected name Spot, got %s\n", addResp.JSON200.Name)
		}
		if addResp.JSON200.Tag == nil || *addResp.JSON200.Tag != "TagOfSpot" {
			failures++
			fmt.Println("   FAIL: tag mismatch")
		}
		fmt.Printf("   OK: created pet ID %d\n", addResp.JSON200.Id)
	}
	if addResp.JSON200 == nil {
		return fmt.Errorf("cannot continue: AddPet returned nil body")
	}
	spotId := addResp.JSON200.Id

	// 2. Find pet by ID
	fmt.Printf("2. Finding pet by ID %d...\n", spotId)
	findResp, err := client.FindPetByIDWithResponse(ctx, spotId)
	if err != nil {
		return fmt.Errorf("FindPetByID failed: %w", err)
	}
	if findResp.StatusCode() != http.StatusOK {
		failures++
		fmt.Printf("   FAIL: expected 200, got %d\n", findResp.StatusCode())
	} else if findResp.JSON200 == nil || findResp.JSON200.Id != spotId {
		failures++
		fmt.Println("   FAIL: pet ID mismatch")
	} else {
		fmt.Println("   OK")
	}

	// 3. Get non-existent pet
	fmt.Println("3. Getting non-existent pet 99999...")
	notFoundResp, err := client.FindPetByIDWithResponse(ctx, 99999)
	if err != nil {
		return fmt.Errorf("FindPetByID failed: %w", err)
	}
	if notFoundResp.StatusCode() != http.StatusNotFound {
		failures++
		fmt.Printf("   FAIL: expected 404, got %d\n", notFoundResp.StatusCode())
	} else {
		fmt.Println("   OK: got 404")
	}

	// 4. Add second pet "Fido"
	fmt.Println("4. Adding pet Fido...")
	tag2 := "TagOfFido"
	addResp2, err := client.AddPetWithResponse(ctx, models.NewPet{
		Name: "Fido",
		Tag:  &tag2,
	})
	if err != nil {
		return fmt.Errorf("AddPet failed: %w", err)
	}
	if addResp2.StatusCode() != http.StatusOK || addResp2.JSON200 == nil {
		failures++
		fmt.Printf("   FAIL: expected 200, got %d\n", addResp2.StatusCode())
	} else {
		fmt.Printf("   OK: created pet ID %d\n", addResp2.JSON200.Id)
	}
	if addResp2.JSON200 == nil {
		return fmt.Errorf("cannot continue: AddPet returned nil body")
	}
	fidoId := addResp2.JSON200.Id

	// 5. List all pets — should have 2
	fmt.Println("5. Listing all pets...")
	listResp, err := client.FindPetsWithResponse(ctx, &models.FindPetsParams{})
	if err != nil {
		return fmt.Errorf("FindPets failed: %w", err)
	}
	if listResp.StatusCode() != http.StatusOK {
		failures++
		fmt.Printf("   FAIL: expected 200, got %d\n", listResp.StatusCode())
	} else if listResp.JSON200 == nil || len(*listResp.JSON200) != 2 {
		failures++
		count := 0
		if listResp.JSON200 != nil {
			count = len(*listResp.JSON200)
		}
		fmt.Printf("   FAIL: expected 2 pets, got %d\n", count)
	} else {
		fmt.Println("   OK: 2 pets")
	}

	// 6. Filter by tag "TagOfFido" — should have 1
	fmt.Println("6. Filtering by tag TagOfFido...")
	tags := []string{"TagOfFido"}
	filterResp, err := client.FindPetsWithResponse(ctx, &models.FindPetsParams{Tags: &tags})
	if err != nil {
		return fmt.Errorf("FindPets failed: %w", err)
	}
	if filterResp.StatusCode() != http.StatusOK {
		failures++
		fmt.Printf("   FAIL: expected 200, got %d\n", filterResp.StatusCode())
	} else if filterResp.JSON200 == nil || len(*filterResp.JSON200) != 1 {
		failures++
		fmt.Println("   FAIL: expected 1 pet")
	} else {
		fmt.Println("   OK: 1 pet")
	}

	// 7. Filter by non-existent tag — should have 0
	fmt.Println("7. Filtering by non-existent tag...")
	noTags := []string{"NotExists"}
	emptyResp, err := client.FindPetsWithResponse(ctx, &models.FindPetsParams{Tags: &noTags})
	if err != nil {
		return fmt.Errorf("FindPets failed: %w", err)
	}
	if emptyResp.StatusCode() != http.StatusOK {
		failures++
		fmt.Printf("   FAIL: expected 200, got %d\n", emptyResp.StatusCode())
	} else if emptyResp.JSON200 == nil || len(*emptyResp.JSON200) != 0 {
		failures++
		fmt.Println("   FAIL: expected 0 pets")
	} else {
		fmt.Println("   OK: 0 pets")
	}

	// 8. Delete non-existent pet — should get 404
	fmt.Println("8. Deleting non-existent pet 99999...")
	delNotFound, err := client.DeletePetWithResponse(ctx, 99999)
	if err != nil {
		return fmt.Errorf("DeletePet failed: %w", err)
	}
	if delNotFound.StatusCode() != http.StatusNotFound {
		failures++
		fmt.Printf("   FAIL: expected 404, got %d\n", delNotFound.StatusCode())
	} else {
		fmt.Println("   OK: got 404")
	}

	// 9. Delete both real pets — should get 204
	fmt.Printf("9. Deleting pet %d...\n", spotId)
	delResp, err := client.DeletePetWithResponse(ctx, spotId)
	if err != nil {
		return fmt.Errorf("DeletePet failed: %w", err)
	}
	if delResp.StatusCode() != http.StatusNoContent {
		failures++
		fmt.Printf("   FAIL: expected 204, got %d\n", delResp.StatusCode())
	} else {
		fmt.Println("   OK: 204")
	}

	fmt.Printf("   Deleting pet %d...\n", fidoId)
	delResp2, err := client.DeletePetWithResponse(ctx, fidoId)
	if err != nil {
		return fmt.Errorf("DeletePet failed: %w", err)
	}
	if delResp2.StatusCode() != http.StatusNoContent {
		failures++
		fmt.Printf("   FAIL: expected 204, got %d\n", delResp2.StatusCode())
	} else {
		fmt.Println("   OK: 204")
	}

	// 10. List all — should have 0
	fmt.Println("10. Listing all pets (should be empty)...")
	finalResp, err := client.FindPetsWithResponse(ctx, &models.FindPetsParams{})
	if err != nil {
		return fmt.Errorf("FindPets failed: %w", err)
	}
	if finalResp.StatusCode() != http.StatusOK {
		failures++
		fmt.Printf("   FAIL: expected 200, got %d\n", finalResp.StatusCode())
	} else if finalResp.JSON200 == nil || len(*finalResp.JSON200) != 0 {
		failures++
		fmt.Println("   FAIL: expected 0 pets")
	} else {
		fmt.Println("   OK: 0 pets")
	}

	if failures > 0 {
		return fmt.Errorf("%d check(s) failed", failures)
	}
	return nil
}
