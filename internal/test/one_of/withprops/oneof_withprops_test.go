package withprops

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSymmetricMarshalling(t *testing.T) {
	expected := TestOne{
		Age:  nil,
		Id:   1,
		Kind: One,
		Name: "one",
	}
	var expectedGeneric Test
	err := expectedGeneric.FromTestOne(expected)
	require.NoError(t, err)
	bytes, err := json.Marshal(expectedGeneric)
	require.NoError(t, err)
	var unmarshalGeneric Test
	err = json.Unmarshal(bytes, &unmarshalGeneric)
	require.NoError(t, err)
	unmarshal, err := unmarshalGeneric.ValueByDiscriminator()
	require.NoError(t, err)
	require.Equal(t, expected, unmarshal.(TestOne))
}
