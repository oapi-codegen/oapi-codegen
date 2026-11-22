package schemasrecursivev1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issue #2542 under v1: the $ref back into the composition is embedded, so
// the recursion closes on the referenced type inside an anonymous item
// struct, and no named item type exists. Compiling is the evidence that
// generation terminated.
var _ Node
var _ Node0
var _ Node1
var _ NodeObject
var _ NodeNestedAllOf
var _ NodeNestedAllOf0
var _ NodeNestedAllOf1
var _ MutualA
var _ MutualB
var _ MutualC
var _ Wrapper
var _ NodeMap

// Whole-document variant: the $ref names no Go type, so the composition
// becomes a named type that refers to itself through the slice.
var _ Tree
var _ Tree_Children_Item

// TestIssue2542ObjectRoundTripV1 checks that the embedded struct carries the
// referenced schema's fields and the sibling's, and that the recursion nests
// to arbitrary depth. Go's encoding/json flattens the embedded NodeObject,
// so the wire format is the same as under v2.
func TestIssue2542ObjectRoundTripV1(t *testing.T) {
	const wire = `{"children":[{"extra":"e","children":[{"extra":"deep"}]}]}`

	var root NodeObject
	require.NoError(t, json.Unmarshal([]byte(wire), &root))
	require.NotNil(t, root.Children)
	require.Len(t, *root.Children, 1)
	child := (*root.Children)[0]
	assert.Equal(t, "e", *child.Extra)
	// Children is promoted from the embedded NodeObject.
	require.NotNil(t, child.Children)
	require.Len(t, *child.Children, 1)
	assert.Equal(t, "deep", *(*child.Children)[0].Extra)

	b, err := json.Marshal(root)
	require.NoError(t, err)
	assert.JSONEq(t, wire, string(b))
}

// TestWholeDocumentRoundTripV1 covers the composition that used to overflow
// the stack under v1: both halves of the allOf are present and the recursion
// nests through the named item type.
func TestWholeDocumentRoundTripV1(t *testing.T) {
	grandchild := Tree_Children_Item{Name: ptr("gc"), Extra: ptr("deep")}
	child := Tree_Children_Item{
		Name:     ptr("c"),
		Extra:    ptr("e"),
		Children: &[]Tree_Children_Item{grandchild},
	}
	root := Tree{Name: ptr("r"), Children: &[]Tree_Children_Item{child}}

	b, err := json.Marshal(root)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"r","children":[{"name":"c","extra":"e","children":[{"name":"gc","extra":"deep"}]}]}`, string(b))

	var back Tree
	require.NoError(t, json.Unmarshal(b, &back))
	require.NotNil(t, back.Children)
	require.Len(t, *back.Children, 1)
	assert.Equal(t, "e", *(*back.Children)[0].Extra)
	inner := (*back.Children)[0].Children
	require.NotNil(t, inner)
	require.Len(t, *inner, 1)
	assert.Equal(t, "deep", *(*inner)[0].Extra)
}

func ptr[T any](v T) *T { return &v }
