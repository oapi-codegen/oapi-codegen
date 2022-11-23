package complex

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetNestedObjects_AddsTypeForChildren(t *testing.T) {
	resp := GetNestedObjectsResponse{
		JSON200: &struct {
			Child *struct {
				IsRequired bool           "json:\"is_required\""
				Names      *[]interface{} "json:\"names,omitempty\""
			} "json:\"child,omitempty\""
			Id *int64 "json:\"id,omitempty\""
		}{
			Child: &struct {
				IsRequired bool           "json:\"is_required\""
				Names      *[]interface{} "json:\"names,omitempty\""
			}{
				IsRequired: false,
				Names:      nil,
			},
			Id: nil,
		},
	}

	require.Equal(t, "*[]string", reflect.TypeOf(resp.JSON200.Child.Names).String())
}

func TestGetWithOperationIdResponse_AddsTypeForProperties(t *testing.T) {
	resp := GetWithOperationIdResponse{
		JSON200: &struct {
			Id    *int64         "json:\"id,omitempty\""
			Names *[]interface{} "json:\"names,omitempty\""
		}{
			Id:    nil,
			Names: nil,
		},
	}

	require.Equal(t, "*[]string", reflect.TypeOf(resp.JSON200.Names).String())
}

func TestGetWithoutOperationIdResponse_AddsTypeForProperties(t *testing.T) {
	resp := GetWithoutOperationIdResponse{
		JSON200: &struct {
			Id    *int64         "json:\"id,omitempty\""
			Names *[]interface{} "json:\"names,omitempty\""
		}{
			Id:    nil,
			Names: nil,
		},
	}

	require.Equal(t, "*[]string", reflect.TypeOf(resp.JSON200.Names).String())
}

func TestGetWithDefault_AddsTypeForProperties(t *testing.T) {
	resp := GetWithDefaultResponse{
		JSONDefault: &struct {
			Id *string `json:"id,omitempty"`
		}{},
	}

	require.Equal(t, "*string", reflect.TypeOf(resp.JSONDefault.Id).String())
}
