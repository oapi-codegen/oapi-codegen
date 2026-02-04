package output

import (
	"encoding/json"
	"testing"
)

// TestAnyOfWithSingleRef verifies that anyOf with a single $ref generates
// correct types that can be used.
// https://github.com/oapi-codegen/oapi-codegen/issues/502
func TestAnyOfWithSingleRef(t *testing.T) {
	// OptionalClaims should be properly generated
	claims := OptionalClaims{
		IDToken:     ptrTo("id-token-value"),
		AccessToken: ptrTo("access-token-value"),
	}

	if *claims.IDToken != "id-token-value" {
		t.Errorf("IDToken = %q, want %q", *claims.IDToken, "id-token-value")
	}
	if *claims.AccessToken != "access-token-value" {
		t.Errorf("AccessToken = %q, want %q", *claims.AccessToken, "access-token-value")
	}
}

func TestApplicationWithAnyOfProperty(t *testing.T) {
	// Application.OptionalClaims is an anyOf with a single ref
	// It should properly wrap OptionalClaims
	app := Application{
		Name: ptrTo("my-app"),
		OptionalClaims: &ApplicationOptionalClaims{
			OptionalClaims: &OptionalClaims{
				IDToken: ptrTo("token"),
			},
		},
	}

	if *app.Name != "my-app" {
		t.Errorf("Name = %q, want %q", *app.Name, "my-app")
	}
	if app.OptionalClaims == nil || app.OptionalClaims.OptionalClaims == nil {
		t.Fatal("OptionalClaims should not be nil")
	}
	if *app.OptionalClaims.OptionalClaims.IDToken != "token" {
		t.Errorf("IDToken = %q, want %q", *app.OptionalClaims.OptionalClaims.IDToken, "token")
	}
}

func TestApplicationJSONRoundTrip(t *testing.T) {
	original := Application{
		Name: ptrTo("test-app"),
		OptionalClaims: &ApplicationOptionalClaims{
			OptionalClaims: &OptionalClaims{
				IDToken:     ptrTo("id"),
				AccessToken: ptrTo("access"),
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Application
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if *decoded.Name != *original.Name {
		t.Errorf("Name mismatch: got %q, want %q", *decoded.Name, *original.Name)
	}
	if decoded.OptionalClaims == nil || decoded.OptionalClaims.OptionalClaims == nil {
		t.Fatal("OptionalClaims should not be nil after round trip")
	}
}

func ptrTo[T any](v T) *T {
	return &v
}
