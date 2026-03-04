package recursive

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// From issue-52: simple recursion (Document → Value → ArrayValue → Value).
func TestSimpleRecursiveTypes(t *testing.T) {
	str := "hello"
	inner := "world"
	doc := Document{
		Fields: &map[string]Value{
			"greeting": {
				StringValue: &str,
			},
			"list": {
				ArrayValue: &ArrayValue{
					{StringValue: &inner},
				},
			},
		},
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	var roundTripped Document
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	assert.NotNil(t, roundTripped.Fields)
	fields := *roundTripped.Fields
	assert.Equal(t, "hello", *fields["greeting"].StringValue)
	assert.Len(t, *fields["list"].ArrayValue, 1)
	assert.Equal(t, "world", *(*fields["list"].ArrayValue)[0].StringValue)
}

func TestDeepRecursiveNesting(t *testing.T) {
	// Build a deeply nested structure: Document → Value → ArrayValue → Value → ArrayValue → Value
	deepStr := "deep"
	midStr := "mid"
	doc := Document{
		Fields: &map[string]Value{
			"level1": {
				ArrayValue: &ArrayValue{
					{
						ArrayValue: &ArrayValue{
							{StringValue: &deepStr},
						},
						StringValue: &midStr,
					},
				},
			},
		},
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	var roundTripped Document
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	fields := *roundTripped.Fields
	level1 := (*fields["level1"].ArrayValue)[0]
	assert.Equal(t, "mid", *level1.StringValue)
	level2 := (*level1.ArrayValue)[0]
	assert.Equal(t, "deep", *level2.StringValue)
}

// From issue-936: deep cyclic oneOf references (FilterPredicate → oneOf → FilterPredicateOp → FilterPredicate).
func TestFilterPredicateWithStringValue(t *testing.T) {
	var fv FilterValue
	err := fv.FromFilterValue1("test-string")
	require.NoError(t, err)

	var fp FilterPredicate
	err = fp.FromFilterValue(fv)
	require.NoError(t, err)

	data, err := json.Marshal(fp)
	require.NoError(t, err)
	assert.Equal(t, `"test-string"`, string(data))

	var roundTripped FilterPredicate
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	gotValue, err := roundTripped.AsFilterValue()
	require.NoError(t, err)
	gotStr, err := gotValue.AsFilterValue1()
	require.NoError(t, err)
	assert.Equal(t, "test-string", gotStr)
}

func TestFilterPredicateWithNumericValue(t *testing.T) {
	var fv FilterValue
	err := fv.FromFilterValue0(42.5)
	require.NoError(t, err)

	var fp FilterPredicate
	err = fp.FromFilterValue(fv)
	require.NoError(t, err)

	data, err := json.Marshal(fp)
	require.NoError(t, err)
	assert.Equal(t, `42.5`, string(data))
}

func TestFilterColumnIncludesInstantiation(t *testing.T) {
	// Verify we can instantiate the full recursive type hierarchy.
	var fv FilterValue
	err := fv.FromFilterValue2(true)
	require.NoError(t, err)

	var fp FilterPredicate
	err = fp.FromFilterValue(fv)
	require.NoError(t, err)

	fci := FilterColumnIncludes{
		Includes: &fp,
	}

	data, err := json.Marshal(fci)
	require.NoError(t, err)

	var roundTripped FilterColumnIncludes
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)
	assert.NotNil(t, roundTripped.Includes)
}
