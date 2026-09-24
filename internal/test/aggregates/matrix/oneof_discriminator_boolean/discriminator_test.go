package matrixoneofdiscriminatorboolean

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBooleanDiscriminator is issue #1619: From* must stamp a JSON boolean,
// so the variant decodes again and ValueByDiscriminator can dispatch on it.
func TestBooleanDiscriminator(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromConfigSsh(ConfigSsh{Host: "h"}))

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"is_http": false, "host": "h"}`, string(b))

	ssh, err := s.AsConfigSsh()
	require.NoError(t, err, "the stamped variant must decode")
	assert.False(t, ssh.IsHttp)

	d, err := s.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "false", d)

	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	assert.IsType(t, ConfigSsh{}, v)

	require.NoError(t, json.Unmarshal([]byte(`{"is_http": true, "host": "h", "port": 80}`), &s))
	v, err = s.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, ConfigHttp{IsHttp: true, Host: "h", Port: ptr(80)}, v)
}

// TestBooleanDiscriminatorWrittenAsString keeps payloads stamped by older
// generated code, which wrote the value as a JSON string, readable.
func TestBooleanDiscriminatorWrittenAsString(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"is_http": "false", "host": "h"}`), &s))
	d, err := s.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "false", d)
}

func ptr[T any](v T) *T { return &v }
