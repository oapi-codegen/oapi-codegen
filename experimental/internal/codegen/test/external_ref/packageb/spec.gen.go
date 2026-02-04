package packageb

// #/components/schemas/ObjectB
type ObjectBSchemaComponent struct {
	Name *string `json:"name,omitempty"`
}

// ApplyDefaults sets default values for fields that are nil.
func (s *ObjectBSchemaComponent) ApplyDefaults() {
}

type ObjectB = ObjectBSchemaComponent
