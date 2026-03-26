package issue2183

import (
	"reflect"
	"strings"
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
		Name string
		// Params defines the query parameters to serialize and deserialize.
		Params []ParamDefinition
		// SpecQuery is the expected raw query string per the OpenAPI spec.
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
			SpecQuery: "search_term=a,b,c%2Cd",
		},
		{
			Name: "explode=true",
			Params: []ParamDefinition{{
				style:     "form",
				explode:   true,
				paramName: "color",
				value:     []string{"blue", "black", "brown"},
			}},
			SpecQuery: "color=blue&color=black&color=brown",
		},
		{
			Name: "multiple params",
			Params: []ParamDefinition{
				{
					style:     "form",
					explode:   false,
					paramName: "color",
					value:     []string{"blue", "black"},
				},
				{
					style:     "form",
					explode:   false,
					paramName: "size",
					value:     []string{"s", "m", "l"},
				},
			},
			SpecQuery: "color=blue,black&size=s,m,l",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			// rawQueryFragments collects pre-encoded query fragments from
			// styled parameters, matching the generated client pattern.
			var rawQueryFragments []string

			for _, param := range tc.Params {

				// following code equivalent to generated client (after fix)
				queryFrag, err := runtime.StyleParamWithOptions(param.style, param.explode, param.paramName, param.value, runtime.StyleParamOptions{ParamLocation: runtime.ParamLocationQuery})
				if err != nil {
					t.Fatal(err)
				}
				rawQueryFragments = append(rawQueryFragments, queryFrag)
			}

			rawQuery := strings.Join(rawQueryFragments, "&")
			t.Logf("client query: %s", rawQuery)
			if tc.SpecQuery != "" && rawQuery != tc.SpecQuery {
				t.Errorf("spec query: expected %q, got %q", tc.SpecQuery, rawQuery)
			}

			// following code equivalent to generated server
			for _, param := range tc.Params {

				dest := reflect.New(reflect.TypeOf(param.value))
				err := runtime.BindRawQueryParameter(param.style, param.explode, true, param.paramName, rawQuery, dest.Interface())
				if err != nil {
					t.Error(err)
				} else if !reflect.DeepEqual(dest.Elem().Interface(), param.value) {
					t.Errorf("expecting %v, got %v", param.value, dest.Elem().Interface())
				}

			}

		})
	}
}
