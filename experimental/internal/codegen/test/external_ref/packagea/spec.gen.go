package packagea

import (
	ext_a5fddf6c "github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen/test/external_ref/packageb"
)

// #/components/schemas/ObjectA
type ObjectASchemaComponent struct {
	Name    *string               `json:"name,omitempty"`
	ObjectB *ext_a5fddf6c.ObjectB `json:"object_b,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ObjectASchemaComponent) ApplyDefaults() {
	if s.ObjectB != nil {
		s.ObjectB.ApplyDefaults()
	}
}

type ObjectA = ObjectASchemaComponent
