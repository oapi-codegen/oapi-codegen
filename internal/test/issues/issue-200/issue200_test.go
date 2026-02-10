package issue200

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDuplicateTypeNamesCompile verifies that when the same name "Bar" is used
// across components/schemas, components/parameters, components/responses,
// components/requestBodies, and components/headers, the codegen produces
// distinct, compilable types with component-based suffixes.
//
// If the auto-rename logic breaks, this test will fail to compile.
func TestDuplicateTypeNamesCompile(t *testing.T) {
	// Schema type: Bar (no suffix, first definition wins)
	_ = Bar{Value: ptr("hello")}

	// Schema types with unique names (no collision)
	_ = Bar2{Value: ptr(float32(1.0))}
	_ = BarParam([]int{1, 2, 3})
	_ = BarParam2([]int{4, 5, 6})

	// Parameter type: BarParameter (was "Bar" in components/parameters)
	_ = BarParameter("query-value")

	// Response type: BarResponse (was "Bar" in components/responses)
	_ = BarResponse{
		Value1: &Bar{Value: ptr("v1")},
		Value2: &Bar2{Value: ptr(float32(2.0))},
		Value3: &BarParam{1},
		Value4: &BarParam2{2},
	}

	// RequestBody type: BarRequestBody (was "Bar" in components/requestBodies)
	_ = BarRequestBody{Value: ptr(42)}

	// Operation-derived types
	_ = PostFooParams{Bar: &Bar{}}
	_ = PostFooJSONBody{Value: ptr(99)}
	_ = PostFooJSONRequestBody{Value: ptr(100)}

	assert.True(t, true, "all duplicate-named types resolved and compiled")
}

func ptr[T any](v T) *T {
	return &v
}
