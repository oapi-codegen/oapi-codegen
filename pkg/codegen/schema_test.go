package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenFieldsFromSchemaDescriptors(t *testing.T) {
	desc := []SchemaDescriptor{
		SchemaDescriptor{
			GoName:   "1name",
			GoType:   "1type",
			Required: true,
			JsonName: "1jname",
		},
		SchemaDescriptor{
			GoName:   "2name",
			GoType:   "2type",
			Required: false,
			JsonName: "2jname",
		},
	}

	res := []string{
		"    1name 1type `json:\"1jname\"`",
		"    2name 2type `json:\"2jname,omitempty\"`",
	}

	f := GenFieldsFromSchemaDescriptors(desc)

	assert.Equal(t, res, f, "Incorrect result")
}
