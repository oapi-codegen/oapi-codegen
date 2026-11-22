package schemasrecursive

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issue #52: recursion via additionalProperties — compile-only.
// The original test called codegen.Generate() to verify no infinite loop;
// here the generated types compiling is sufficient evidence.
var _ Document
var _ Value
var _ ArrayValue

// issue #936: cyclic oneOf — compile-only.
// The original test verified generation succeeds; compilation confirms it.
var _ FilterColumnIncludes
var _ FilterPredicate
var _ FilterPredicateOp
var _ FilterPredicateRangeOp
var _ FilterRangeValue
var _ FilterValue

// issue #1373: recursive $ref via allOf — compile-only.
// The original test verified generation succeeds; compilation confirms it.
var _ RecursiveObject
var _ NonRecursiveObject

// issue #2542: an allOf that composes a schema whose own body contains that
// allOf. Generation used to overflow the stack. The composition now becomes a
// named type that refers back to itself, which is how Go expresses this and
// how the generator already represents every non-recursive allOf: flattened
// and concrete, with no union wrapper.
var _ Node
var _ Node0
var _ Node1
var _ Node_1_Children_Item

// Object variant: no union anywhere, just a struct that contains a slice of
// itself.
var _ NodeObject
var _ NodeObject_Children_Item

// Transitive variant: the self-$ref hides behind an allOf nested inside
// another allOf member.
var _ NodeNestedAllOf
var _ NodeNestedAllOf_1_Children_Item

// Mutual variant: the cycle never passes through the component being
// generated.
var _ MutualA
var _ MutualB
var _ MutualC

// Bystander variant: the cycle is reached from a component outside it.
var _ Wrapper
var _ Wrapper_N_Children_Item

// additionalProperties variant: the cycle closes through a map, at the
// component root. items and additionalProperties reuse their parent's path,
// so this sits at path length 1.
var _ NodeMap
var _ NodeMap_AdditionalProperties

// TestIssue2542ObjectRoundTrip checks that the composed type carries both
// halves of the allOf — the referenced schema's fields and the sibling's —
// and that the recursion nests to arbitrary depth.
func TestIssue2542ObjectRoundTrip(t *testing.T) {
	grandchild := NodeObject_Children_Item{Extra: ptr("deep")}
	child := NodeObject_Children_Item{
		Extra:    ptr("e"),
		Children: &[]NodeObject_Children_Item{grandchild},
	}
	root := NodeObject{Children: &[]NodeObject_Children_Item{child}}

	b, err := json.Marshal(root)
	require.NoError(t, err)
	assert.JSONEq(t, `{"children":[{"extra":"e","children":[{"extra":"deep"}]}]}`, string(b))

	var back NodeObject
	require.NoError(t, json.Unmarshal(b, &back))
	require.NotNil(t, back.Children)
	require.Len(t, *back.Children, 1)
	assert.Equal(t, "e", *(*back.Children)[0].Extra)
	inner := (*back.Children)[0].Children
	require.NotNil(t, inner)
	require.Len(t, *inner, 1)
	assert.Equal(t, "deep", *(*inner)[0].Extra)
}

// TestIssue2542UnionRoundTrip covers the shape from the issue itself. Node is
// an anyOf, so the composed type is still a union — over Node's own branches
// regenerated at this position, which is what a non-recursive allOf over a
// union already produces.
func TestIssue2542UnionRoundTrip(t *testing.T) {
	var leaf Node_1_Children_Item
	require.NoError(t, leaf.FromNode1Children0(Node1Children0{Leaf: ptr("l")}))
	leaf.Extra = ptr("e")

	var root Node
	require.NoError(t, root.FromNode1(Node1{Children: &[]Node_1_Children_Item{leaf}}))

	b, err := json.Marshal(root)
	require.NoError(t, err)
	assert.JSONEq(t, `{"children":[{"leaf":"l","extra":"e"}]}`, string(b))

	var back Node
	require.NoError(t, json.Unmarshal(b, &back))
	n1, err := back.AsNode1()
	require.NoError(t, err)
	require.NotNil(t, n1.Children)
	require.Len(t, *n1.Children, 1)
	assert.Equal(t, "e", *(*n1.Children)[0].Extra)
}

func ptr[T any](v T) *T { return &v }
