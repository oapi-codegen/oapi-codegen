package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const constraintsTestSpec = `
openapi: 3.0.0
info:
  title: Constraints Test
  version: 1.0.0
components:
  schemas:
    # Number type with minimum and maximum
    Age:
      type: integer
      minimum: 0
      maximum: 150
      default: 25

    # Number with exclusive bounds
    Temperature:
      type: number
      format: float
      minimum: -273.15
      maximum: 1000.0
      exclusiveMinimum: true
      exclusiveMaximum: false

    # String type with length constraints
    Username:
      type: string
      minLength: 3
      maxLength: 20
      default: "user"

    # Array type with item constraints
    Tags:
      type: array
      items:
        type: string
      minItems: 1
      maxItems: 10
      default: ["tag1"]

    # Boolean with default
    IsActive:
      type: boolean
      default: true

    # Integer with various formats
    Port:
      type: integer
      format: int32
      minimum: 1
      maximum: 65535
      default: 8080

    # Number without constraints (should not generate constants)
    Price:
      type: number

    # String without constraints
    Description:
      type: string
`

func TestGenerateConstraints(t *testing.T) {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(constraintsTestSpec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName: "testconstraints",
		Generate: GenerateOptions{
			Models: true,
		},
		OutputOptions: OutputOptions{
			SkipPrune: true,
		},
	}

	// Run full code generation which will include constraints
	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	// Test Age constraints
	assert.Contains(t, code, "AgeMinimum")
	assert.Contains(t, code, "AgeMaximum")
	assert.Contains(t, code, "AgeDefault")
	assert.Contains(t, code, "int = 0")
	assert.Contains(t, code, "int = 150")
	assert.Contains(t, code, "int = 25")

	// Test Temperature constraints
	assert.Contains(t, code, "TemperatureMinimum")
	assert.Contains(t, code, "TemperatureMaximum")
	assert.Contains(t, code, "float32 = -273.15")
	assert.Contains(t, code, "float32 = 1000")

	// Test Username constraints
	assert.Contains(t, code, "UsernameMinLength")
	assert.Contains(t, code, "UsernameMaxLength")
	assert.Contains(t, code, "UsernameDefault")
	assert.Contains(t, code, "uint64 = 3")
	assert.Contains(t, code, "uint64 = 20")
	assert.Contains(t, code, `string = "user"`)

	// Test Tags constraints
	assert.Contains(t, code, "TagsMinItems")
	assert.Contains(t, code, "TagsMaxItems")
	assert.Contains(t, code, "uint64 = 1")
	assert.Contains(t, code, "uint64 = 10")

	// Test IsActive default
	assert.Contains(t, code, "IsActiveDefault")
	assert.Contains(t, code, "bool = true")

	// Test Port constraints
	assert.Contains(t, code, "PortMinimum")
	assert.Contains(t, code, "PortMaximum")
	assert.Contains(t, code, "PortDefault")
	assert.Contains(t, code, "int32 = 1")
	assert.Contains(t, code, "int32 = 65535")
	assert.Contains(t, code, "int32 = 8080")

	// Test that types without constraints don't generate constants
	assert.NotContains(t, code, "PriceMinimum")
	assert.NotContains(t, code, "PriceMaximum")
	assert.NotContains(t, code, "DescriptionMinLength")
}

func TestConstraintDefinitionExtraction(t *testing.T) {
	tests := []struct {
		name          string
		spec          string
		typeName      string
		wantMin       *float64
		wantMax       *float64
		wantDefault   interface{}
		wantMinLength *uint64
		wantMaxLength *uint64
		wantMinItems  *uint64
		wantMaxItems  *uint64
	}{
		{
			name: "integer with all numeric constraints",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    TestInt:
      type: integer
      minimum: 10
      maximum: 100
      default: 50
`,
			typeName:    "TestInt",
			wantMin:     ptrFloat64(10),
			wantMax:     ptrFloat64(100),
			wantDefault: float64(50),
		},
		{
			name: "string with length constraints",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    TestString:
      type: string
      minLength: 5
      maxLength: 50
      default: "hello"
`,
			typeName:      "TestString",
			wantMinLength: ptrUint64(5),
			wantMaxLength: ptrUint64(50),
			wantDefault:   "hello",
		},
		{
			name: "array with item constraints",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    TestArray:
      type: array
      items:
        type: string
      minItems: 2
      maxItems: 20
`,
			typeName:     "TestArray",
			wantMinItems: ptrUint64(2),
			wantMaxItems: ptrUint64(20),
		},
		{
			name: "boolean with default only",
			spec: `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
components:
  schemas:
    TestBool:
      type: boolean
      default: false
`,
			typeName:    "TestBool",
			wantDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			swagger, err := loader.LoadFromData([]byte(tt.spec))
			require.NoError(t, err)

			types, err := GenerateTypesForSchemas(nil, swagger.Components.Schemas, nil)
			require.NoError(t, err)
			require.Len(t, types, 1)

			tp := types[0]
			require.Equal(t, tt.typeName, tp.TypeName)
			require.NotNil(t, tp.Schema.OAPISchema)

			schema := tp.Schema.OAPISchema

			if tt.wantMin != nil {
				require.NotNil(t, schema.Min)
				assert.Equal(t, *tt.wantMin, *schema.Min)
			}

			if tt.wantMax != nil {
				require.NotNil(t, schema.Max)
				assert.Equal(t, *tt.wantMax, *schema.Max)
			}

			if tt.wantDefault != nil {
				require.NotNil(t, schema.Default)
				assert.Equal(t, tt.wantDefault, schema.Default)
			}

			if tt.wantMinLength != nil {
				assert.Equal(t, *tt.wantMinLength, schema.MinLength)
			}

			if tt.wantMaxLength != nil {
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, *tt.wantMaxLength, *schema.MaxLength)
			}

			if tt.wantMinItems != nil {
				assert.Equal(t, *tt.wantMinItems, schema.MinItems)
			}

			if tt.wantMaxItems != nil {
				require.NotNil(t, schema.MaxItems)
				assert.Equal(t, *tt.wantMaxItems, *schema.MaxItems)
			}
		})
	}
}

func TestInlineParameterConstraints(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: Parameter Constraints Test
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      parameters:
        - in: query
          name: limit
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - in: query
          name: search
          schema:
            type: string
            minLength: 3
            maxLength: 50
        - in: query
          name: offset
          schema:
            type: integer
            minimum: 0
      responses:
        '200':
          description: Success
`

	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName: "testparams",
		Generate: GenerateOptions{
			Models: true,
		},
		OutputOptions: OutputOptions{
			SkipPrune: true,
		},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	// Test limit parameter constraints
	assert.Contains(t, code, "GetUsersLimitMinimum")
	assert.Contains(t, code, "GetUsersLimitMaximum")
	assert.Contains(t, code, "GetUsersLimitDefault")
	assert.Contains(t, code, "int = 1")
	assert.Contains(t, code, "int = 100")
	assert.Contains(t, code, "int = 20")

	// Test search parameter constraints
	assert.Contains(t, code, "GetUsersSearchMinLength")
	assert.Contains(t, code, "GetUsersSearchMaxLength")
	assert.Contains(t, code, "uint64 = 3")
	assert.Contains(t, code, "uint64 = 50")

	// Test offset parameter constraints (only minimum)
	assert.Contains(t, code, "GetUsersOffsetMinimum")
	assert.Contains(t, code, "int = 0")
}

// Helper functions
func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrUint64(v uint64) *uint64 {
	return &v
}
