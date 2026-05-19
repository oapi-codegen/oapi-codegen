package enabled

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// When `skip-response-body-getters` is unset (default), the generated response
// type must expose `GetBody()` and a typed getter for each response field.
func TestResponseBodyGettersGenerated(t *testing.T) {
	respType := reflect.TypeOf(GetThingResponse{})

	t.Run("GetBody returns the raw body", func(t *testing.T) {
		m, ok := respType.MethodByName("GetBody")
		require.True(t, ok, "GetBody method should be generated")

		// signature: func (r GetThingResponse) GetBody() []byte
		require.Equal(t, 1, m.Type.NumOut())
		assert.Equal(t, reflect.TypeOf([]byte(nil)), m.Type.Out(0))

		body := []byte("hello")
		r := GetThingResponse{Body: body, HTTPResponse: &http.Response{}}
		assert.Equal(t, body, r.GetBody())
	})

	t.Run("typed JSON200 getter returns the decoded payload", func(t *testing.T) {
		m, ok := respType.MethodByName("GetJSON200")
		require.True(t, ok, "GetJSON200 method should be generated")

		// signature: func (r GetThingResponse) GetJSON200() *Thing
		require.Equal(t, 1, m.Type.NumOut())
		assert.Equal(t, reflect.TypeOf(&Thing{}), m.Type.Out(0))

		thing := &Thing{Id: "1", Name: "rock"}
		r := GetThingResponse{JSON200: thing}
		assert.Same(t, thing, r.GetJSON200())
	})

	t.Run("typed JSONDefault getter returns the decoded error payload", func(t *testing.T) {
		m, ok := respType.MethodByName("GetJSONDefault")
		require.True(t, ok, "GetJSONDefault method should be generated")

		require.Equal(t, 1, m.Type.NumOut())
		assert.Equal(t, reflect.TypeOf(&Error{}), m.Type.Out(0))
	})
}
