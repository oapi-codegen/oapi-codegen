package wholefileunion_test

import (
	"encoding/json"
	"testing"

	wholefileunion "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/external/whole_file_union"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResponseBodyUnion round-trips a response body whose oneOf branches are
// both whole-file refs, through the accessors named after the branch paths.
func TestResponseBodyUnion(t *testing.T) {
	var body wholefileunion.GetItems200JSONResponseBody
	require.NoError(t, body.FromGetItems200JSONResponseBody0(wholefileunion.GetItems200JSONResponseBody0{Name: ptr("alice")}))

	b, err := json.Marshal(body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"alice"}`, string(b))

	var decoded wholefileunion.GetItems200JSONResponseBody
	require.NoError(t, json.Unmarshal([]byte(`{"id":7}`), &decoded))
	group, err := decoded.AsGetItems200JSONResponseBody1()
	require.NoError(t, err)
	require.NotNil(t, group.Id)
	assert.Equal(t, 7, *group.Id)
}

// TestMixedUnion checks that a whole-file branch next to an inline branch
// gets its own type, just like the inline one.
func TestMixedUnion(t *testing.T) {
	var body wholefileunion.GetMixed200JSONResponseBody
	require.NoError(t, body.FromGetMixed200JSONResponseBody0(wholefileunion.GetMixed200JSONResponseBody0{Name: ptr("alice")}))
	require.NoError(t, body.MergeGetMixed200JSONResponseBody1(wholefileunion.GetMixed200JSONResponseBody1{Label: ptr("admin")}))

	b, err := json.Marshal(body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"alice","label":"admin"}`, string(b))
}

// TestComponentUnion round-trips a component schema whose oneOf branches are
// whole-file refs.
func TestComponentUnion(t *testing.T) {
	var item wholefileunion.Item
	require.NoError(t, json.Unmarshal([]byte(`{"name":"alice"}`), &item))
	user, err := item.AsItem0()
	require.NoError(t, err)
	require.NotNil(t, user.Name)
	assert.Equal(t, "alice", *user.Name)
}

// TestPropertyUnion round-trips an object property whose oneOf branches are
// whole-file refs.
func TestPropertyUnion(t *testing.T) {
	var member wholefileunion.Holder_Member
	require.NoError(t, member.FromHolderMember1(wholefileunion.HolderMember1{Id: ptr(7)}))

	b, err := json.Marshal(wholefileunion.Holder{Member: &member})
	require.NoError(t, err)
	assert.JSONEq(t, `{"member":{"id":7}}`, string(b))
}

func ptr[T any](v T) *T { return &v }
