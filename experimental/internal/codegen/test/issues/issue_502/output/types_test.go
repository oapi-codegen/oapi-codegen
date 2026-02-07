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
	// Application.OptionalClaims is an anyOf with a single ref + nullable: true
	// It should be Nullable[ApplicationOptionalClaims]
	app := Application{
		Name: ptrTo("my-app"),
		OptionalClaims: NewNullableWithValue(ApplicationOptionalClaims{
			OptionalClaims: &OptionalClaims{
				IDToken: ptrTo("token"),
			},
		}),
	}

	if *app.Name != "my-app" {
		t.Errorf("Name = %q, want %q", *app.Name, "my-app")
	}
	if !app.OptionalClaims.IsSpecified() {
		t.Fatal("OptionalClaims should be specified")
	}
	optClaims := app.OptionalClaims.MustGet()
	if optClaims.OptionalClaims == nil {
		t.Fatal("OptionalClaims.OptionalClaims should not be nil")
	}
	if *optClaims.OptionalClaims.IDToken != "token" {
		t.Errorf("IDToken = %q, want %q", *optClaims.OptionalClaims.IDToken, "token")
	}
}

func TestApplicationJSONRoundTrip(t *testing.T) {
	original := Application{
		Name: ptrTo("test-app"),
		OptionalClaims: NewNullableWithValue(ApplicationOptionalClaims{
			OptionalClaims: &OptionalClaims{
				IDToken:     ptrTo("id"),
				AccessToken: ptrTo("access"),
			},
		}),
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
	if !decoded.OptionalClaims.IsSpecified() {
		t.Fatal("OptionalClaims should be specified after round trip")
	}
	optClaims := decoded.OptionalClaims.MustGet()
	if optClaims.OptionalClaims == nil {
		t.Fatal("OptionalClaims.OptionalClaims should not be nil after round trip")
	}
}

func TestApplicationNullOptionalClaims(t *testing.T) {
	// Test with explicitly null optional claims
	app := Application{
		Name:           ptrTo("null-test-app"),
		OptionalClaims: NewNullNullable[ApplicationOptionalClaims](),
	}

	if !app.OptionalClaims.IsNull() {
		t.Error("OptionalClaims should be null")
	}

	// Should marshal as null
	data, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	t.Logf("Marshaled with null optionalClaims: %s", string(data))
}

func ptrTo[T any](v T) *T {
	return &v
}
