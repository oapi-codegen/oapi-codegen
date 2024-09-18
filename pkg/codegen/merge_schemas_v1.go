package codegen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func mergeSchemasV1(allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	var outSchema Schema
	for _, schemaOrRef := range allOf {
		ref := schemaOrRef.Ref

		var refType string
		var err error
		if IsGoTypeReference(ref) {
			refType, err = RefPathToGoType(ref)
			if err != nil {
				return Schema{}, fmt.Errorf("error converting reference path to a go type: %w", err)
			}
		}

		schema, err := GenerateGoSchema(schemaOrRef, path)
		if err != nil {
			return Schema{}, fmt.Errorf("error generating Go schema in allOf: %w", err)
		}
		schema.RefType = refType

		for _, p := range schema.Properties {
			err = outSchema.AddProperty(p)
			if err != nil {
				return Schema{}, fmt.Errorf("error merging properties: %w", err)
			}
		}

		if schema.HasAdditionalProperties {
			if outSchema.HasAdditionalProperties {
				// Both this schema, and the aggregate schema have additional
				// properties, they must match.
				if schema.AdditionalPropertiesType.TypeDecl() != outSchema.AdditionalPropertiesType.TypeDecl() {
					return Schema{}, errors.New("additional properties in allOf have incompatible types")
				}
			} else {
				// We're switching from having no additional properties to having
				// them
				outSchema.HasAdditionalProperties = true
				outSchema.AdditionalPropertiesType = schema.AdditionalPropertiesType
			}
		}
	}

	// Now, we generate the struct which merges together all the fields.
	var err error
	outSchema.GoType, err = GenStructFromAllOf(allOf, path)
	if err != nil {
		return Schema{}, fmt.Errorf("unable to generate aggregate type for AllOf: %w", err)
	}
	return outSchema, nil
}

// GenStructFromAllOf generates an object that is the union of the objects in the
// input array. In the case of Ref objects, we use an embedded struct, otherwise,
// we inline the fields.
func GenStructFromAllOf(allOf []*openapi3.SchemaRef, path []string) (string, error) {
	// Start out with struct {
	objectParts := []string{"struct {"}
	for _, schemaOrRef := range allOf {
		ref := schemaOrRef.Ref
		if IsGoTypeReference(ref) {
			// We have a referenced type, we will generate an inlined struct
			// member.
			// struct {
			//   InlinedMember
			//   ...
			// }
			goType, err := RefPathToGoType(ref)
			if err != nil {
				return "", err
			}
			objectParts = append(objectParts,
				fmt.Sprintf("   // Embedded struct due to allOf(%s)", ref))
			objectParts = append(objectParts,
				fmt.Sprintf("   %s `yaml:\",inline\"`", goType))
		} else {
			// Inline all the fields from the schema into the output struct,
			// just like in the simple case of generating an object.
			goSchema, err := GenerateGoSchema(schemaOrRef, path)
			if err != nil {
				return "", err
			}
			objectParts = append(objectParts, "   // Embedded fields due to inline allOf schema")
			objectParts = append(objectParts, GenFieldsFromProperties(goSchema.Properties)...)

			if goSchema.HasAdditionalProperties {
				addPropsType := goSchema.AdditionalPropertiesType.GoType
				if goSchema.AdditionalPropertiesType.RefType != "" {
					addPropsType = goSchema.AdditionalPropertiesType.RefType
				}

				additionalPropertiesPart := fmt.Sprintf("AdditionalProperties map[string]%s `json:\"-\"`", addPropsType)
				if !StringInArray(additionalPropertiesPart, objectParts) {
					objectParts = append(objectParts, additionalPropertiesPart)
				}
			}
		}
	}
	objectParts = append(objectParts, "}")
	return strings.Join(objectParts, "\n"), nil
}
