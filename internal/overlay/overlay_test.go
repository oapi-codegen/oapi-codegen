package overlay

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const header = `overlay: 1.0.0
info:
  title: test
  version: 1.0.0
`

func TestParseAndValidate(t *testing.T) {
	tests := []struct {
		name     string
		overlay  string
		parseErr string
		wantErrs []string
	}{
		{
			name:    "valid",
			overlay: header + "extends: https://example.com/openapi.yaml\nactions:\n  - target: $.info\n    update: {x: 1}\n",
		},
		{
			name:    "a later 1.0 patch version",
			overlay: "overlay: 1.0.7\ninfo: {title: t, version: v}\nactions: [{target: $, remove: false}]\n",
		},
		{
			name:    "numeric info fields",
			overlay: "overlay: 1.0.0\ninfo: {title: 1, version: 2}\nactions: [{target: $, update: {}}]\n",
		},
		{
			name:     "everything missing",
			overlay:  "x-extension: true\n",
			wantErrs: []string{"overlay version must be defined", "overlay info title must be defined", "overlay info version must be defined", "overlay must define at least one action"},
		},
		{
			name:     "unsupported version",
			overlay:  "overlay: 1.1.0\ninfo: {title: t, version: v}\nactions: [{target: $, update: {}}]\n",
			wantErrs: []string{`overlay version "1.1.0" is not supported: it must be 1.0.x`},
		},
		{
			name:     "not a version",
			overlay:  "overlay: one\ninfo: {title: t, version: v}\nactions: [{target: $, update: {}}]\n",
			wantErrs: []string{`overlay version "one" is not supported`},
		},
		{
			name:     "empty actions",
			overlay:  header + "actions: []\n",
			wantErrs: []string{"overlay must define at least one action"},
		},
		{
			name:     "action without a target",
			overlay:  header + "actions:\n  - update: {x: 1}\n  - target: ''\n",
			wantErrs: []string{"overlay action at index 0 target must be defined", "overlay action at index 1 target must be defined"},
		},
		{
			name:     "remove and update",
			overlay:  header + "actions:\n  - target: $.info\n    remove: true\n    update: {x: 1}\n",
			wantErrs: []string{"overlay action at index 0 should not both set remove and define update"},
		},
		{
			name:     "remove and a null update",
			overlay:  header + "actions:\n  - target: $.info\n    remove: true\n    update: null\n",
			wantErrs: []string{"should not both set remove and define update"},
		},
		{
			name:     "extends is not a URL",
			overlay:  header + "extends: \"http://[::1\"\nactions: [{target: $, update: {}}]\n",
			wantErrs: []string{"overlay extends must be a valid URL"},
		},
		{
			name:     "wrongly typed fields of an action",
			overlay:  header + "actions:\n  - target: 1\n    remove: yes please\n  - not an action\n",
			wantErrs: []string{"overlay action at index 0 target must be a string", "overlay action at index 0 remove must be a boolean", "overlay action at index 1 must be a mapping"},
		},
		{
			name:     "overlay is a number",
			overlay:  "overlay: 1.0\ninfo: {title: t, version: v}\nactions: [{target: $, update: {}}]\n",
			parseErr: "overlay version must be a string",
		},
		{
			name:     "not a mapping",
			overlay:  "- a\n- b\n",
			parseErr: "the document is not a mapping",
		},
		{
			name:     "invalid YAML",
			overlay:  "overlay: [\n",
			parseErr: "failed to parse overlay",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := Parse([]byte(tt.overlay))
			if tt.parseErr != "" {
				require.ErrorContains(t, err, tt.parseErr)
				return
			}
			require.NoError(t, err)
			err = o.Validate()
			if len(tt.wantErrs) == 0 {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			for _, want := range tt.wantErrs {
				assert.ErrorContains(t, err, want)
			}
		})
	}
}

func TestParseFields(t *testing.T) {
	o, err := Parse([]byte(header + `extends: https://example.com/openapi.yaml
actions:
  - target: $.paths
    description: add a response
    update:
      /ping:
        get:
          responses:
            200:
              description: ok
  - target: $.paths['/x']
    remove: true
  - target: $.info.description
    update: null
`))
	require.NoError(t, err)
	require.NoError(t, o.Validate())
	assert.Equal(t, "1.0.0", o.Version)
	assert.Equal(t, Info{Title: "test", Version: "1.0.0"}, o.Info)
	assert.Equal(t, "https://example.com/openapi.yaml", o.Extends)
	require.Len(t, o.Actions, 3)

	assert.Equal(t, "$.paths", o.Actions[0].Target)
	assert.Equal(t, "add a response", o.Actions[0].Description)
	assert.True(t, o.Actions[0].HasUpdate)
	// A YAML key that reads as a number becomes a string.
	assert.Equal(t, map[string]any{"/ping": map[string]any{"get": map[string]any{"responses": map[string]any{"200": map[string]any{"description": "ok"}}}}}, o.Actions[0].Update)

	assert.True(t, o.Actions[1].Remove)
	assert.False(t, o.Actions[1].HasUpdate)

	assert.True(t, o.Actions[2].HasUpdate, "update: null is an update")
	assert.Nil(t, o.Actions[2].Update)
}

func mustDecode(t *testing.T, s string) any {
	t.Helper()
	v, err := DecodeJSON([]byte(s))
	require.NoError(t, err)
	return v
}

func mustParse(t *testing.T, s string) *Overlay {
	t.Helper()
	o, err := Parse([]byte(header + s))
	require.NoError(t, err)
	require.NoError(t, o.Validate())
	return o
}

func toJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func TestApply(t *testing.T) {
	const doc = `{
		"info": {"title": "api", "version": "1.0", "x-list": [1, 2]},
		"tags": ["a", "b"],
		"paths": {
			"/a": {"get": {"tags": ["internal"], "x-internal": true}, "post": {"tags": ["public"]}},
			"/b": {"x-internal": true, "get": {}}
		},
		"items": [{"n": 0}, {"n": 1, "drop": true}, {"n": 2}, {"n": 3, "drop": true}, {"n": 4, "drop": true}]
	}`
	tests := []struct {
		name    string
		actions string
		want    string
	}{
		{
			name: "merge into a mapping",
			actions: `actions:
  - target: $.info
    update:
      title: renamed
      description: added
      x-list: [3]
`,
			want: `{"info":{"description":"added","title":"renamed","version":"1.0","x-list":[1,2,3]}}`,
		},
		{
			name: "structural update of the root",
			actions: `actions:
  - target: $
    update:
      info:
        x-extra: {nested: true}
      tags: [c]
`,
			want: `{"info":{"title":"api","version":"1.0","x-extra":{"nested":true},"x-list":[1,2]},"tags":["a","b","c"]}`,
		},
		{
			name: "a nested value of another kind is replaced",
			actions: `actions:
  - target: $
    update:
      info: {x-list: {not: a list}}
      tags: plain
`,
			want: `{"info":{"title":"api","version":"1.0","x-list":{"not":"a list"}},"tags":"plain"}`,
		},
		{
			name: "append a value to a sequence",
			actions: `actions:
  - target: $.tags
    update: c
  - target: $.info['x-list']
    update: {k: v}
`,
			want: `{"info":{"title":"api","version":"1.0","x-list":[1,2,{"k":"v"}]},"tags":["a","b","c"]}`,
		},
		{
			name: "append the elements of a sequence to a sequence",
			actions: `actions:
  - target: $.tags
    update: [c, d]
`,
			want: `{"tags":["a","b","c","d"]}`,
		},
		{
			name: "replace a scalar",
			actions: `actions:
  - target: $.info.title
    update: renamed
  - target: $.info.version
    update: {now: an object}
`,
			want: `{"info":{"title":"renamed","version":{"now":"an object"},"x-list":[1,2]}}`,
		},
		{
			name: "update every match of a filter, each with its own copy",
			actions: `actions:
  - target: $.paths.*[?@.tags]
    update:
      x-seen: [1]
  - target: $.paths['/a'].get['x-seen']
    update: 2
`,
			want: `{"paths":{"/a":{"get":{"tags":["internal"],"x-internal":true,"x-seen":[1,2]},"post":{"tags":["public"],"x-seen":[1]}},"/b":{"get":{},"x-internal":true}}}`,
		},
		{
			name: "remove a member",
			actions: `actions:
  - target: $.info.version
    remove: true
`,
			want: `{"info":{"title":"api","x-list":[1,2]}}`,
		},
		{
			name: "remove by filter at several levels",
			actions: `actions:
  - target: $.paths[?@['x-internal']]
    remove: true
  - target: $.paths.*[?@['x-internal']]
    remove: true
`,
			want: `{"paths":{"/a":{"post":{"tags":["public"]}}}}`,
		},
		{
			name: "remove by tag",
			actions: `actions:
  - target: $.paths.*[?@.tags[?@ == 'internal']]
    remove: true
`,
			want: `{"paths":{"/a":{"post":{"tags":["public"]}},"/b":{"get":{},"x-internal":true}}}`,
		},
		{
			name: "remove several elements of one sequence",
			actions: `actions:
  - target: $.items[?@.drop]
    remove: true
`,
			want: `{"items":[{"n":0},{"n":2}]}`,
		},
		{
			name: "remove elements selected out of order and twice",
			actions: `actions:
  - target: $.items[4,0,4,2]
    remove: true
`,
			want: `{"items":[{"n":1,"drop":true},{"n":3,"drop":true}]}`,
		},
		{
			name: "remove a node and a node inside it",
			actions: `actions:
  - target: $..[?@['x-internal']]
    remove: true
`,
			want: `{"paths":{"/a":{"post":{"tags":["public"]}}}}`,
		},
		{
			name: "actions apply in order",
			actions: `actions:
  - target: $.info
    update: {x-new: 1}
  - target: $.info['x-new']
    remove: true
  - target: $.info
    update: {title: last}
`,
			want: `{"info":{"title":"last","version":"1.0","x-list":[1,2]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := mustParse(t, tt.actions)
			got, err := o.Apply(mustDecode(t, doc), Options{Strict: true})
			require.NoError(t, err)
			// Compare only the top-level members the expectation lists, so
			// each case can focus on what it changes.
			want := mustDecode(t, tt.want).(map[string]any)
			gotMap := got.(map[string]any)
			for k := range want {
				assert.JSONEq(t, toJSON(t, want[k]), toJSON(t, gotMap[k]), k)
			}
		})
	}
}

func TestApplyReplacesTheRoot(t *testing.T) {
	o := mustParse(t, "actions:\n  - target: $\n    update: [1]\n")
	got, err := o.Apply(mustDecode(t, `{"a": 1}`), Options{})
	require.NoError(t, err)
	assert.Equal(t, []any{json.Number("1")}, got)
}

func TestApplyRemoveRoot(t *testing.T) {
	o := mustParse(t, "actions:\n  - target: $\n    remove: true\n")
	_, err := o.Apply(mustDecode(t, `{"a": 1}`), Options{})
	require.ErrorContains(t, err, "cannot remove the document root ($)")
}

func TestApplyStrict(t *testing.T) {
	o := mustParse(t, `actions:
  - target: $.missing
    description: first miss
    update: {x: 1}
  - target: $.info
    update: {x-applied: true}
  - target: $.paths[?@.nothing]
    remove: true
`)
	doc := `{"info": {"title": "t"}, "paths": {"/a": {}}}`

	got, err := o.Apply(mustDecode(t, doc), Options{})
	require.NoError(t, err, "non-strict mode ignores targets that match nothing")
	assert.JSONEq(t, `{"info": {"title": "t", "x-applied": true}, "paths": {"/a": {}}}`, toJSON(t, got))

	_, err = o.Apply(mustDecode(t, doc), Options{Strict: true})
	var unmatched *UnmatchedTargetsError
	require.True(t, errors.As(err, &unmatched), "got %v", err)
	assert.Equal(t, []string{
		`action at index 0 (target "$.missing", description "first miss")`,
		`action at index 2 (target "$.paths[?@.nothing]")`,
	}, unmatched.Actions)
	assert.ErrorContains(t, err, "did not match anything")
}

func TestApplyRegexFunctionsOnMissingMembers(t *testing.T) {
	o := mustParse(t, `actions:
  - target: $.paths[?search(@.missing, '.*')]
    remove: true
  - target: $.paths[?match(@.missing, '.*')]
    remove: true
  - target: $.paths[?search(@.operationId, '^int')]
    update: {x-matched: true}
`)
	doc := `{"paths": {"/a": {"operationId": "internalPing"}, "/b": {}}}`
	got, err := o.Apply(mustDecode(t, doc), Options{})
	require.NoError(t, err)
	assert.JSONEq(t, `{"paths": {"/a": {"operationId": "internalPing", "x-matched": true}, "/b": {}}}`, toJSON(t, got))

	_, err = o.Apply(mustDecode(t, doc), Options{Strict: true})
	var unmatched *UnmatchedTargetsError
	require.True(t, errors.As(err, &unmatched), "got %v", err)
	assert.Len(t, unmatched.Actions, 2, "neither search() nor match() selects a node without the member")
}

func TestApplyKeepsNumbers(t *testing.T) {
	o := mustParse(t, `actions:
  - target: $.components.schemas[?@.minimum == 1.5]
    update: {x-matched: 0.1}
  - target: $.components.schemas.N.maximum
    update: 12345678901234567890
`)
	doc := `{"components": {"schemas": {"N": {"minimum": 1.50, "multipleOf": 0.10, "big": 123456789012345678901234567890, "exp": 1e400, "maximum": 1, "example": {"v": 1.50}}}}}`
	got, err := o.Apply(mustDecode(t, doc), Options{Strict: true})
	require.NoError(t, err)
	assert.Equal(t,
		`{"components":{"schemas":{"N":{"big":123456789012345678901234567890,"example":{"v":1.50},"exp":1e400,"maximum":12345678901234567890,"minimum":1.50,"multipleOf":0.10,"x-matched":0.1}}}}`,
		toJSON(t, got))
}

func TestApplyLegacyJSONPath(t *testing.T) {
	doc := `{"paths": {"/a": {"get": {"x-internal": true}, "put": {}}}}`
	o := mustParse(t, `actions:
  - target: $.paths.*.*[?(@.x-internal)]
    remove: true
`)
	got, err := o.Apply(mustDecode(t, doc), Options{Strict: true, LegacyJSONPath: true})
	require.NoError(t, err)
	assert.JSONEq(t, `{"paths": {"/a": {"put": {}}}}`, toJSON(t, got))
}

func TestApplyTargetErrors(t *testing.T) {
	doc := `{"paths": {}}`
	tests := []struct {
		name     string
		target   string
		legacy   bool
		wantErrs []string
		notWant  string
	}{
		{
			name:   "legacy target in RFC mode suggests a translation",
			target: "$.paths.*.*[?(@.x-internal)]",
			wantErrs: []string{
				`overlay action at index 0 (target "$.paths.*.*[?(@.x-internal)]")`,
				"is not valid RFC 9535 JSONPath",
				`the RFC 9535 equivalent is probably "$.paths.*[?@['x-internal']]"`,
				"set compatibility.overlays-use-yaml-jsonpath: true",
			},
		},
		{
			name:     "an invalid target without a translation still names the option",
			target:   "$.paths[",
			wantErrs: []string{"is not valid RFC 9535 JSONPath", "set compatibility.overlays-use-yaml-jsonpath: true"},
			notWant:  "probably",
		},
		{
			name:     "RFC target in legacy mode",
			target:   "$.paths[?@.x]",
			legacy:   true,
			wantErrs: []string{"compatibility.overlays-use-yaml-jsonpath: true is set", "the target is valid RFC 9535 JSONPath"},
		},
		{
			name:     "untranslatable legacy target",
			target:   "$.paths['/a'].get[?(@.x)]",
			legacy:   true,
			wantErrs: []string{"cannot translate", "RFC 9535 cannot test a node by its own key"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := mustParse(t, "actions:\n  - target: \""+tt.target+"\"\n    remove: true\n")
			_, err := o.Apply(mustDecode(t, doc), Options{LegacyJSONPath: tt.legacy})
			require.Error(t, err)
			for _, want := range tt.wantErrs {
				assert.ErrorContains(t, err, want)
			}
			if tt.notWant != "" {
				assert.NotContains(t, err.Error(), tt.notWant)
			}
		})
	}
}
