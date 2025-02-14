package codegen

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-test/deep"
)

func TestAllOf(t *testing.T) {
	suite := []testCase{
		// just properties
		makeTestCase(
			"when 1 level deep, it merges schemas",
			`[
				{ "properties": { "a_foo": { "type": "string" } } },
				{ "properties": { "b_foo": { "type": "string" } } }
			]`,
			`{ "a_foo": { "type": "string" }, "b_foo": { "type": "string" } }`,
			"",
			nil,
		),
		makeTestCase(
			"when 1 level deep with 1 schema, it returns single schema",
			`[
				{ "properties": { "a_foo": { "type": "string" } } }
			]`,
			`{ "a_foo": { "type": "string" } }`,
			"",
			nil,
		),
		makeTestCase(
			"when 3 levels deep, it merges schemas",
			`[
				{
					"allOf": [
						{
							"allOf": [
								{"properties": { "a0_foo": { "type": "string" } }},
								{"properties": { "a1_foo": { "type": "string" } }}
							]
						}
					]
				},
				{
					"allOf": [
						{
							"allOf": [
								{"properties": { "b0_foo": { "type": "string" } }},
								{"properties": { "b1_foo": { "type": "string" } }}
							]
						}
					]
				}
			]`,
			`{
				"a0_foo": { "type": "string" },
				"a1_foo": { "type": "string" },
				"b0_foo": { "type": "string" },
				"b1_foo": { "type": "string" }
			}`,
			"",
			nil,
		),
		makeTestCase(
			"when 3 levels deep with single schema, it flattens to single schema",
			`[
				{
					"allOf": [
						{
							"allOf": [
								{"properties": { "a0_foo": { "type": "string" } }},
								{"properties": { "a1_foo": { "type": "string" } }}
							]
						}
					]
				}
			]`,
			`{
				"a0_foo": { "type": "string" },
				"a1_foo": { "type": "string" }
			}`,
			"",
			nil,
		),
		// include single oneOf
		makeTestCase(
			"when 1 level deep with oneOf, it merges schemas",
			`[
				{
					"properties": { "a_foo": { "type": "string" } },
					"oneOf": [
						{ "properties": { "a_foo_one_of_0": { "type": "string" } }},
						{ "properties": { "a_foo_one_of_1": { "type": "string" } }}
					]
				},
				{ "properties": { "b_foo": { "type": "string" } } }
			]`,
			`{ "a_foo": { "type": "string" }, "b_foo": { "type": "string" } }`,
			`[
				{ "properties": { "a_foo_one_of_0": { "type": "string" } }},
				{ "properties": { "a_foo_one_of_1": { "type": "string" } }}
			]`,
			nil,
		),
		makeTestCase(
			"when 3 levels deep with oneOf at leaf, it merges schemas",
			`[
				{
					"allOf": [
						{
							"allOf": [
								{
									"properties": { "a0_foo": { "type": "string" } },
									"oneOf": [
										{ "properties": { "a_foo_one_of_0": { "type": "string" } }},
										{ "properties": { "a_foo_one_of_1": { "type": "string" } }}
									]
								},
								{ "properties": { "a1_foo": { "type": "string" } } }
							]
						}
					]
				},
				{
					"allOf": [
						{
							"allOf": [
								{"properties": { "b0_foo": { "type": "string" } }},
								{"properties": { "b1_foo": { "type": "string" } }}
							]
						}
					]
				}
			]`,
			`{
				"a0_foo": { "type": "string" },
				"a1_foo": { "type": "string" },
				"b0_foo": { "type": "string" },
				"b1_foo": { "type": "string" }
			}`,
			`[
				{ "properties": { "a_foo_one_of_0": { "type": "string" } }},
				{ "properties": { "a_foo_one_of_1": { "type": "string" } }}
			]`,
			nil,
		),
		makeTestCase(
			"when 3 levels deep with oneOf in the middle, it merges schemas",
			`[
				{
					"allOf": [
						{
							"oneOf": [
								{ "properties": { "a_foo_one_of_0": { "type": "string" } }},
								{ "properties": { "a_foo_one_of_1": { "type": "string" } }}
							]
						},
						{
							"allOf": [
								{"properties": { "a0_foo": { "type": "string" } }},
								{"properties": { "a1_foo": { "type": "string" } }}
							]
						}
					]
				},
				{
					"allOf": [
						{
							"allOf": [
								{"properties": { "b0_foo": { "type": "string" } }},
								{"properties": { "b1_foo": { "type": "string" } }}
							]
						}
					]
				}
			]`,
			`{
				"a0_foo": { "type": "string" },
				"a1_foo": { "type": "string" },
				"b0_foo": { "type": "string" },
				"b1_foo": { "type": "string" }
			}`,
			`[
				{ "properties": { "a_foo_one_of_0": { "type": "string" } }},
				{ "properties": { "a_foo_one_of_1": { "type": "string" } }}
			]`,
			nil,
		),
		makeTestCase(
			"with multi-level oneOf, it merges schemas",
			`[
				{
					"allOf": [
						{
							"allOf": [
								{"properties": { "a0_foo": { "type": "string" } }},
								{"properties": { "a1_foo": { "type": "string" } }}
							]
						},
						{
							"oneOf": [
								{
									"allOf": [
										{"properties": { "a_foo_one_of_a_0": { "type": "string" } }},
										{"properties": { "a_foo_one_of_a_1": { "type": "string" } }}
									]
								},
								{
									"allOf": [
										{"properties": { "a_foo_one_of_b_0": { "type": "string" } }},
										{"properties": { "a_foo_one_of_b_1": { "type": "string" } }}
									]
								}
							]
						}
					]
				},
				{
					"allOf": [
						{
							"allOf": [
								{"properties": { "b0_foo": { "type": "string" } }},
								{"properties": { "b1_foo": { "type": "string" } }}
							]
						}
					]
				}
			]`,
			`{
				"a0_foo": { "type": "string" },
				"a1_foo": { "type": "string" },
				"b0_foo": { "type": "string" },
				"b1_foo": { "type": "string" }
			}`,
			`[
				{
					"properties": {
						"a_foo_one_of_a_0": { "type": "string" },
						"a_foo_one_of_a_1": { "type": "string" }
					}
				},
				{
					"properties": {
						"a_foo_one_of_b_0": { "type": "string" },
						"a_foo_one_of_b_1": { "type": "string" }
					}
				}
			]`,
			nil,
		),
		// other test cases
		makeTestCase(
			"should preserve nested oneOf",
			`[
				{
					"allOf": [
						{
							"allOf": [
								{
									"oneOf": [
										{"properties": { "a_foo_one_of_0": { "type": "string" } }},
										{"properties": { "a_foo_one_of_1": { "type": "string" } }}
									]
								}
							]
						}, 
						{ "properties": { "a_foo": { "type": "string" } } }
					]
				}
			]`,
			`{
				"a_foo": { "type": "string" }
			}`,
			`[
				{ "properties": { "a_foo_one_of_0": { "type": "string" } }},
				{ "properties": { "a_foo_one_of_1": { "type": "string" } }}
			]`,
			nil,
		),
	}

	runSuite(t, suite)
}

func runSuite(t *testing.T, suite []testCase) {
	for _, test := range suite {
		t.Run(test.name, func(t *testing.T) {
			result, err := MergeSchemas(test.inputAllOf, make([]string, 0))

			if diff := deep.Equal(err, test.expectedError); diff != nil {
				var errString = "Error validation failed. diff:"
				for _, diffItem := range diff {
					errString += "\n\t" + diffItem
				}
				t.Fatal(errString)
				return
			}
			if err != nil {
				return
			}

			// result.OAPISchema is an intermediate schema (it's currently the schema after recursing
			// AllOf, but but before recursing Properties/Enums/everything else that needs recursing),
			// so we need to pull merged props off the "codegen" version
			// NOTE: normally result.OAPISchema would be the "original" schema when calling
			//       `GenerateGoSchema`, but we are testing `MergeSchemas` directly here
			err = validateProperties(&test.expectedProperties, &result)
			if err != nil {
				t.Fatal(err)
				return
			}

			// similar to above, the OneOfs are after the first pass or merging, so we extract
			// the properties for comparison
			// TODO: current tests only care about properties, but we should also test other fields
			if len(result.UnionElements) != len(test.expectedOneOf) {
				t.Fatalf("Expected %d oneOfs, got %d", len(test.expectedOneOf), len(result.UnionElements))
				return
			}
			for i, oneOf := range result.UnionElements {
				typeIdx := slices.IndexFunc(result.AdditionalTypes, func(t TypeDefinition) bool { return t.TypeName == string(oneOf) })
				if typeIdx == -1 {
					t.Fatalf("Expected oneOf %d to have type %s, but it was not found", i, oneOf)
					return
				}
				oneOfType := result.AdditionalTypes[typeIdx]
				expectedOneOf := test.expectedOneOf[i]
				// see notes on validateProperties above
				err = validateProperties(&expectedOneOf.Value.Properties, &oneOfType.Schema)
				if err != nil {
					t.Fatal(err)
					return
				}
			}
		})
	}
}

func validateProperties(expectedProperties *openapi3.Schemas, result *Schema) error {
	props := openapi3.Schemas{}
	for _, prop := range result.Properties {
		props[prop.JsonFieldName] = openapi3.NewSchemaRef("", prop.Schema.OAPISchema)
	}
	if diff := deep.Equal(props, *expectedProperties); diff != nil {
		var errString = "Properties validation failed. diff:"
		for _, diffItem := range diff {
			errString += "\n\t" + diffItem
		}
		return errors.New(errString)
	}

	return nil
}

type testCase struct {
	name               string
	inputAllOf         openapi3.SchemaRefs
	expectedProperties openapi3.Schemas
	expectedOneOf      openapi3.SchemaRefs
	expectedError      error
}

func makeTestCase(
	name string,
	inputAllOfSchema string,
	expectedPropertiesSchema string,
	expectedOneOfSchema string,
	expectedError error,
) testCase {
	inputAllOf := openapi3.SchemaRefs{}
	if err := json.Unmarshal([]byte(inputAllOfSchema), &inputAllOf); err != nil {
		panic(err)
	}

	var expectedProperties openapi3.Schemas
	if len(expectedPropertiesSchema) > 0 {
		if err := json.Unmarshal([]byte(expectedPropertiesSchema), &expectedProperties); err != nil {
			panic(err)
		}
	}

	var expectedOneOf openapi3.SchemaRefs = nil
	if len(expectedOneOfSchema) > 0 {
		expectedOneOf = openapi3.SchemaRefs{}
		if err := json.Unmarshal([]byte(expectedOneOfSchema), &expectedOneOf); err != nil {
			panic(err)
		}
	}

	return testCase{
		name:               name,
		inputAllOf:         inputAllOf,
		expectedProperties: expectedProperties,
		expectedOneOf:      expectedOneOf,
		expectedError:      expectedError,
	}
}
