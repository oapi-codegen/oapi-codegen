package externalref

import (
	"testing"

	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen/test/external_ref/packagea"
	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen/test/external_ref/packageb"
)

// TestCrossPackageReferences verifies that types from external packages
// can be used correctly in the Container type.
func TestCrossPackageReferences(t *testing.T) {
	// Create objects from each package
	name := "test-name"
	b := &packageb.ObjectB{
		Name: &name,
	}

	a := &packagea.ObjectA{
		Name:    &name,
		ObjectB: b,
	}

	// Create container that references both
	container := Container{
		ObjectA: a,
		ObjectB: b,
	}

	// Verify the structure
	if container.ObjectA == nil {
		t.Error("ObjectA should not be nil")
	}
	if container.ObjectB == nil {
		t.Error("ObjectB should not be nil")
	}
	if *container.ObjectA.Name != "test-name" {
		t.Errorf("ObjectA.Name = %q, want %q", *container.ObjectA.Name, "test-name")
	}
	if *container.ObjectB.Name != "test-name" {
		t.Errorf("ObjectB.Name = %q, want %q", *container.ObjectB.Name, "test-name")
	}

	// Verify nested reference in ObjectA
	if container.ObjectA.ObjectB == nil {
		t.Error("ObjectA.ObjectB should not be nil")
	}
	if *container.ObjectA.ObjectB.Name != "test-name" {
		t.Errorf("ObjectA.ObjectB.Name = %q, want %q", *container.ObjectA.ObjectB.Name, "test-name")
	}
}

// TestApplyDefaults verifies that ApplyDefaults works across package boundaries.
func TestApplyDefaults(t *testing.T) {
	container := Container{
		ObjectA: &packagea.ObjectA{},
		ObjectB: &packageb.ObjectB{},
	}

	// Should not panic when calling ApplyDefaults across packages
	container.ApplyDefaults()
}
