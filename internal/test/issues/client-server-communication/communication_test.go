package client_server_communication

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/oapi-codegen/runtime"
)

func TestQueryCommunication(t *testing.T) {

	type ParamDefinition struct {
		style     string
		explode   bool
		paramName string
		value     any
	}

	testCases := []struct {
		Name      string
		Params    []ParamDefinition
		SpecQuery string
	}{
		{
			Name: "explode=false",
			Params: []ParamDefinition{{
				style:     "form",
				explode:   false,
				paramName: "color",
				value:     []string{"blue", "black", "brown"},
			}},
			// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#style-examples
			SpecQuery: "color=blue,black,brown",
		},
		{
			Name: "explode=false;commas",
			Params: []ParamDefinition{{
				style:     "form",
				explode:   false,
				paramName: "search_term",
				value:     []string{"a", "b", "c,d"},
			}},
			SpecQuery: "search_terms=a,b,c%2Cd",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			queryValues := url.Values{}

			for _, param := range tc.Params {

				// following code equivalent to generated client
				queryFrag, err := runtime.StyleParamWithLocation(param.style, param.explode, param.paramName, runtime.ParamLocationQuery, param.value)
				if err != nil {
					t.Fatal(err)
				}
				parsed, err := url.ParseQuery(queryFrag)
				if err != nil {
					t.Fatal(err)
				}
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

			rawQuery := queryValues.Encode()
			t.Logf("client query: %s", rawQuery)
			if tc.SpecQuery != "" && rawQuery != tc.SpecQuery {
				t.Errorf("spec query: expected %q, got %q", tc.SpecQuery, rawQuery)
			}

			serverValues, err := url.ParseQuery(rawQuery)
			if err != nil {
				t.Fatal(err)
			}

			// following code equivalent to generated server
			for _, param := range tc.Params {

				dest := reflect.New(reflect.TypeOf(param.value))
				err := runtime.BindQueryParameter(param.style, param.explode, true, param.paramName, serverValues, dest.Interface())
				if err != nil {
					t.Error(err)
				} else if !reflect.DeepEqual(dest.Elem().Interface(), param.value) {
					t.Errorf("expecting %v, got %v", param.value, dest.Elem().Interface())
				}

			}

		})
	}
}
