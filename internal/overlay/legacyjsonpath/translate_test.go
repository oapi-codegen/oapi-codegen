package legacyjsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theory/jsonpath"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		name   string
		legacy string
		want   string
	}{
		// Paths that are already RFC 9535 stay as they are.
		{"root", "$", "$"},
		{"plain names", "$.components.schemas.Pong", "$.components.schemas.Pong"},
		{"bracketed name", "$.paths['/ping'].get", "$.paths['/ping'].get"},
		{"double-quoted name", `$.paths["/ping"].get`, "$.paths['/ping'].get"},
		{"name union", `$.paths['/a', "/b"]`, "$.paths['/a','/b']"},
		{"index and slice", "$.servers[0].variables[1:3]", "$.servers[0].variables[1:3]"},
		{"wildcards", "$.paths.*[*]", "$.paths.*[*]"},
		{"recursive descent", "$..description", "$..description"},
		{"recursive descent with bracket", "$..['x-go-name']", "$..['x-go-name']"},
		{"recursive wildcard", "$..*", "$..*"},

		// Names that RFC 9535 does not allow in shorthand form.
		{"name with dash", "$.components.schemas.Pong.x-internal", "$.components.schemas.Pong['x-internal']"},
		{"name with slash", "$.paths./ping.get.responses.200.content.application/json", "$.paths['/ping'].get.responses['200'].content['application/json']"},
		{"name starting with a digit", "$.responses.4XX", "$.responses['4XX']"},
		{"recursive descent to a dashed name", "$..x-internal", "$..['x-internal']"},
		{"quote in name", `$["it's"]`, `$['it\'s']`},
		{"backslash in name", `$['a\\b']`, `$['a\\b']`},

		// An implicit root.
		{"implicit root", "paths.*", "$.paths.*"},
		{"implicit root with dash", "x-tags[0]", "$['x-tags'][0]"},
		{"implicit root with bracket", "['paths']", "$['paths']"},

		// Filters move up one level when a wildcard precedes them.
		{"filter after dot wildcard", "$.paths.*.*[?(@.x-internal)]", "$.paths.*[?@['x-internal']]"},
		{"filter after bracket wildcard", "$.paths[*][?(@.x-internal)]", "$.paths[?@['x-internal']]"},
		{"filter on the path items", "$.paths.*[?(@.x-internal)]", "$.paths[?@['x-internal']]"},
		{"negated filter", "$.paths.*.*[?(!@.servers)]", "$.paths.*[?!@.servers]"},
		{"filter on schema properties", "$.components.schemas.*.*.*[?(@.x-internal)]", "$.components.schemas.*.*[?@['x-internal']]"},
		{"filter after recursive descent", "$.paths.*.*.requestBody.*.*.schema..[?(@.x-internal)]", "$.paths.*.*.requestBody.*.*.schema..[?@['x-internal']]"},
		{"filter after recursive wildcard", "$..*[?(@.x-internal)]", "$..[?@['x-internal']]"},
		{"filter with a path after it", "$.paths.*.*[?(@.deprecated == true)].responses", "$.paths.*[?@.deprecated == true].responses"},
		{"nested filter", "$.paths.*.*[?(@.parameters[*][?(@.in == 'query')])]", "$.paths.*[?@.parameters[?@.in == 'query']]"},

		// Filter expressions.
		{"string comparison", "$.paths.*.*[?(@.operationId == 'ping')]", "$.paths.*[?@.operationId == 'ping']"},
		{"double-quoted string", `$.paths.*.*[?(@.operationId == "it's")]`, `$.paths.*[?@.operationId == 'it\'s']`},
		{"string with backslash", `$.paths.*.*[?(@.description == 'a\b')]`, `$.paths.*[?@.description == 'a\\b']`},
		{"numbers", "$.a.*[?(@.x >= .5 && @.y < 10 || @.z != -3)]", "$.a[?@.x >= 0.5 && @.y < 10 || @.z != -3]"},
		{"literals", "$.a.*[?(@.x == true || @.y == null)]", "$.a[?@.x == true || @.y == null]"},
		{"literal first", "$.a.*[?('ping' == @.operationId)]", "$.a[?'ping' == @.operationId]"},
		{"parentheses", "$.a.*[?((@.x || @.y) && @.z)]", "$.a[?(@.x || @.y) && @.z]"},
		{"negated comparison", "$.a.*[?(!@.x == 1)]", "$.a[?!(@.x == 1)]"},
		{"negated negation", "$.a.*[?(!!@.x)]", "$.a[?!(!@.x)]"},
		{"root query in filter", "$.a.*[?(@.v == $.info.version)]", "$.a[?@.v == $.info.version]"},
		{"current node", "$.tags[*][?(@ == 'internal')]", "$.tags[?@ == 'internal']"},
		{"bracketed name in filter", "$.a.*[?(@['x-go-name'])]", "$.a[?@['x-go-name']]"},

		// =~ becomes search().
		{"regex", "$.paths.*.*[?(@.operationId =~ /^internal/)]", "$.paths.*[?search(@.operationId, '^internal')]"},
		{"regex with escapes", `$.paths.*[?(@.summary =~ /a\/b\d'/)]`, `$.paths[?search(@.summary, 'a/b\\d\'')]`},
		{"negated regex", "$.paths.*.*[?(!@.operationId =~ /x/)]", "$.paths.*[?!search(@.operationId, 'x')]"},
		{"regex over a wildcard", "$.paths.*.*[?(@.tags[*] =~ /^int/)]", "$.paths.*[?@.tags[?search(@, '^int')]]"},

		// Comparisons over several nodes become nested filters.
		{"tags comparison", "$.paths.*.*[?(@.tags[*] == 'internal')]", "$.paths.*[?@.tags[?@ == 'internal']]"},
		{"dot wildcard comparison", "$.a.*[?(@.tags.* != 'x')]", "$.a[?@.tags[?@ != 'x']]"},
		{"wildcard with suffix", "$.a.*[?(@.servers[*].url == 'http://x')]", "$.a[?@.servers[?@.url == 'http://x']]"},
		{"literal before wildcard", "$.a.*[?('internal' == @.tags[*])]", "$.a[?@.tags[?'internal' == @]]"},
		{"root query over wildcard", "$.a.*[?($.tags[*].name == 'x')]", "$.a[?$.tags[?@.name == 'x']]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Translate(tt.legacy)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			_, err = jsonpath.Parse(got)
			assert.NoError(t, err, "the translation must be valid RFC 9535")
		})
	}
}

func TestTranslateErrors(t *testing.T) {
	tests := []struct {
		name    string
		legacy  string
		wantErr string
	}{
		{"empty", "", "empty target"},
		{"filter on a named node", "$.components.schemas.Pong[?(@.x-internal)]", "applies to $.components.schemas.Pong itself"},
		{"filter on a named node suggests [*]", "$.paths['/ping'].get.parameters[?(@.in == 'query')]", "write [*] before the filter"},
		{"filter on the root", "$[?(@.openapi)]", "applies to $ itself"},
		{"filter after recursive descent to a name", "$..schema[?(@.x)]", "applies to $..schema itself"},
		{"filter after a filter", "$.a.*[?(@.x)][?(@.y)]", "itself"},
		{"property name", "$.paths.*~", "~ property-name operator"},
		{"bracket property name", "$.paths['/a']~", "~ property-name operator"},
		{"subscript property name", "$.paths[*]~", "~ property-name operator"},
		{"regex flags", "$.a.*[?(@.x =~ /abc/i)]", "flags are not supported"},
		{"invalid regex", "$.a.*[?(@.x =~ /(/)]", "invalid regular expression"},
		{"unterminated regex", "$.a.*[?(@.x =~ /abc)]", "unmatched regular expression"},
		{"unterminated filter", "$.a.*[?(@.x", "missing end of filter"},
		{"unterminated bracket", "$.a[0", "unmatched ["},
		{"unterminated quote", "$['a", "unmatched '"},
		{"bad escape", `$['a\n']`, "unsupported escape sequence"},
		{"bad index", "$.a[x]", "invalid array index"},
		{"empty name", "$.a..", "child name or array access or filter missing"},
		{"missing name", "$.a.", "child name missing"},
		{"literal as filter", "$.a.*[?('x')]", "a literal on its own"},
		{"regex on literal", "$.a.*[?('x' =~ /x/)]", "literal cannot be matched"},
		{"two multi-node queries", "$.a.*[?(@.x[*] == @.y[*])]", "each select several nodes"},
		{"multi-node query against @", "$.a.*[?(@.x[*] == @.y)]", "relative to @"},
		{"recursive descent in comparison", "$.a.*[?(@..x == 1)]", "single wildcard"},
		{"two wildcards in comparison", "$.a.*[?(@.x[*].y[*] == 1)]", "single wildcard"},
		{"garbage after path", "$.a b", "unexpected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Translate(tt.legacy)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

// TestTranslateAnyVersusAll pins down the known semantic gap: the previous
// dialect compared every node a wildcard selected, and the translation
// compares any of them.
func TestTranslateAnyVersusAll(t *testing.T) {
	got, err := Translate("$.paths.*.*[?(@.tags[*] == 'internal')]")
	require.NoError(t, err)
	path, err := jsonpath.Parse(got)
	require.NoError(t, err)

	doc := map[string]any{
		"paths": map[string]any{
			"/only-internal": map[string]any{
				"get": map[string]any{"tags": []any{"internal"}},
			},
			"/mixed": map[string]any{
				"get": map[string]any{"tags": []any{"internal", "public"}},
			},
			"/public": map[string]any{
				"get": map[string]any{"tags": []any{"public"}},
			},
		},
	}
	var selected []string
	for _, n := range path.SelectLocated(doc) {
		selected = append(selected, n.Path.String())
	}
	// The previous dialect selected only /only-internal, as /mixed has a tag
	// that is not "internal". The translation selects every operation that
	// has the tag.
	assert.ElementsMatch(t, []string{
		"$['paths']['/only-internal']['get']",
		"$['paths']['/mixed']['get']",
	}, selected)
}

// TestTranslateExamples covers the targets in the README and in
// examples/overlay before they were rewritten in RFC 9535.
func TestTranslateExamples(t *testing.T) {
	for legacy, want := range map[string]string{
		"$":                          "$",
		"$.paths.*.*[?(!@.servers)]": "$.paths.*[?!@.servers]",
		"$.components.schemas.Pong":  "$.components.schemas.Pong",
		"$.paths['/ping'].get":       "$.paths['/ping'].get",
		"$.components.schemas.Pong.properties.seed":             "$.components.schemas.Pong.properties.seed",
		"$.components.schemas.*.*.*[?(@.x-internal)]":           "$.components.schemas.*.*[?@['x-internal']]",
		"$.paths.*.*.requestBody.*.*.schema..[?(@.x-internal)]": "$.paths.*.*.requestBody.*.*.schema..[?@['x-internal']]",
		"$.paths.*.*.responses.*.*.*.schema..[?(@.x-internal)]": "$.paths.*.*.responses.*.*.*.schema..[?@['x-internal']]",
		"$.paths.*[?(@.x-internal)]":                            "$.paths[?@['x-internal']]",
		"$.paths.*.*[?(@.x-internal)]":                          "$.paths.*[?@['x-internal']]",
		"$.paths.*.*[?(@.tags[*] == 'internal')]":               "$.paths.*[?@.tags[?@ == 'internal']]",
	} {
		got, err := Translate(legacy)
		require.NoError(t, err, legacy)
		assert.Equal(t, want, got, legacy)
	}
}
