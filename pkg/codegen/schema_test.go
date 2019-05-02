package codegen

import "testing"

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

	if len(f) != 2 {
		t.Error("Incorrect len")
		return
	}

	for i, s := range f {
		if s != res[i] {
			t.Error("Incorrect result")
		}
	}
}
