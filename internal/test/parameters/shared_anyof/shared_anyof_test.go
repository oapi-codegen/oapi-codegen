package sharedanyof

import (
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSharedAnyOfParamTypesDeclaredOnce is the issue-2090 regression guard: a
// path-level anyOf parameter shared by several methods, and a component
// parameter referenced from several paths, each have their union member types
// declared exactly once. The blank declarations below only build because each
// type exists (and is not redeclared); the generated package failing to build
// as part of the test module would also catch a regression.
func TestSharedAnyOfParamTypesDeclaredOnce(t *testing.T) {
	var (
		_ Id0       // path parameter, shared by get/post/delete
		_ Id1       //
		_ Whid0     // webhook query parameter
		_ Cbid0     // callback query parameter
		_ WidgetId0 // component parameter, $ref'd from two paths
		_ WidgetId1 //
	)

	// The component parameter's union type carries its accessors once.
	u := openapi_types.UUID{}
	var w WidgetId
	require.NoError(t, w.FromWidgetId0(u))
	got, err := w.AsWidgetId0()
	require.NoError(t, err)
	assert.Equal(t, u, got)
}
