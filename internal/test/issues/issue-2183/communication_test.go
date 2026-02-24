package issue2183

import (
	"net/url"
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
		// SkipServerRoundTrip skips the server-side deserialization check.
		// This is needed for cases where values contain the delimiter character
		// (comma for form/explode=false), because Go's url.ParseQuery decodes
		// %2C to "," before BindQueryParameter can distinguish delimiters from
		// literal commas. Fixing this requires changes to the runtime library.
		SkipServerRoundTrip bool
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
			SpecQuery:           "search_term=a,b,c%2Cd",
			SkipServerRoundTrip: true,
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
				queryFragments, err := runtime.StyleParamWithLocation(param.style, param.explode, param.paramName, runtime.ParamLocationQuery, param.value)
				if err != nil {
					t.Fatal(err)
				}
				rawQueryFragments = append(rawQueryFragments, queryFragments)
			}

			rawQuery := strings.Join(rawQueryFragments, "&")
			t.Logf("client query: %s", rawQuery)
			if tc.SpecQuery != "" && rawQuery != tc.SpecQuery {
				t.Errorf("spec query: expected %q, got %q", tc.SpecQuery, rawQuery)
			}

			if tc.SkipServerRoundTrip {
				return
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
