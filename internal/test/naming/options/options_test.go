package namingoptions

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The names used here are the test: with the naming options off, none of them
// would exist, and this file would not compile.

func TestEnumValuesKeepUnderscores(t *testing.T) {
	assert.Equal(t, LogLevel("LOG_LEVEL_DEBUG"), LOG_LEVEL_DEBUG)
	assert.Equal(t, LogLevel("log_level_info"), Log_Level_Info)
	// A hyphen is still a separator that is removed.
	assert.Equal(t, LogLevel("LOG-LEVEL-WARN"), LOGLEVELWARN)
	// A leading underscore keeps its word, so the constant stays exported.
	assert.Equal(t, LogLevel("_hidden"), Underscore_Hidden)

	assert.True(t, LOG_LEVEL_DEBUG.Valid())
	assert.False(t, LogLevel("LOGLEVELDEBUG").Valid())
}

func TestExplicitEnumNamesVerbatim(t *testing.T) {
	// Color and Channel share the value "b", so their constants are prefixed
	// with the type name, except for the names the spec states.
	assert.Equal(t, Color("r"), COLOR_RED)
	assert.Equal(t, Color("g"), Color_green)
	assert.Equal(t, Color("b"), ColorB)

	assert.Equal(t, Channel("a"), CHANNEL_ALPHA)
	assert.Equal(t, Channel("b"), ChannelB)

	assert.True(t, Color_green.Valid())
}

func TestPropertyNames(t *testing.T) {
	id, upper, hardCoded := "an id", "upper", "hard coded"
	level := Log_Level_Info
	thing := Thing{
		// Leading underscores are dropped.
		Id:             &id,
		UPPERSNAKECASE: &upper,
		// x-go-name, as written.
		HARD_CODED_ID: &hardCoded,
		Level:         &level,
	}

	buf, err := json.Marshal(thing)
	require.NoError(t, err)
	assert.JSONEq(t, `{"_id":"an id","_UPPER_SNAKE_CASE":"upper","hardCodedId":"hard coded","level":"log_level_info"}`, string(buf))

	var decoded Thing
	require.NoError(t, json.Unmarshal(buf, &decoded))
	assert.Equal(t, thing, decoded)
}

func TestParameterNameVerbatim(t *testing.T) {
	pageSize := 10
	params := ListThingsParams{Page_Size: &pageSize}

	req, err := NewListThingsRequest("https://example.com", &params)
	require.NoError(t, err)
	assert.Equal(t, "10", req.URL.Query().Get("page_size"))
}
