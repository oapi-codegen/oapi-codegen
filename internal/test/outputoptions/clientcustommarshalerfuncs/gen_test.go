package clientcustommarshalerfuncs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetJSONUnmarshal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"name": "custom_name"}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	c, err := NewClientWithResponses(server.URL)
	assert.NoError(t, err)

	resp, err := c.GetClientWithResponse(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "custom_name", resp.JSON200.Name)

	SetJSONUnmarshal(func(data []byte, v any) error {
		if err := json.Unmarshal(data, v); err != nil {
			return err
		}

		resp, ok := v.(*ClientType)
		assert.True(t, ok)
		resp.Name = "test_unmarshal"

		return nil
	})

	resp, err = c.GetClientWithResponse(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test_unmarshal", resp.JSON200.Name)
}

func TestSetJSONMarshal(t *testing.T) {
	var isCustomJSON bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isCustomJSON {
			assert.Equal(t, `/contentObject/{"Id":111,"IsAdmin":false}`, r.URL.Path)
			return
		}

		assert.Equal(t, `/contentObject/{"Id":100500,"IsAdmin":true}`, r.URL.Path)
	}))
	defer server.Close()

	c, err := NewClientWithResponses(server.URL)
	assert.NoError(t, err)

	_, err = c.GetContentObjectWithResponse(context.Background(), ComplexObject{
		Id:      100500,
		IsAdmin: true,
	})
	assert.NoError(t, err)

	isCustomJSON = true

	SetJSONMarshal(func(v any) ([]byte, error) {
		obj, ok := v.(ComplexObject)
		assert.True(t, ok)

		obj.Id = 111
		obj.IsAdmin = false

		return json.Marshal(obj)
	})

	_, err = c.GetContentObjectWithResponse(context.Background(), ComplexObject{
		Id:      100500,
		IsAdmin: true,
	})
	assert.NoError(t, err)
}
