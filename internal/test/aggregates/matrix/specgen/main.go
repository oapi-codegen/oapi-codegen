// Command specgen writes the aggregate test matrix: for every shape described
// in ../shapes/*.yaml, and every schema-merging-behavior version the shape
// lists, it writes a package ../<version>/<shape>/ holding an OpenAPI spec that
// places the shape's subject schema at every position we generate code for, an
// oapi-codegen config pinned to that version, a doc.go with the go:generate
// directive, and a round-trip test harness.
//
// Run it through `go generate` in the matrix directory. A new shape needs a
// second `go generate` pass, because go generate lists packages before specgen
// creates the new one.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"

	"go.yaml.in/yaml/v3"
)

// Positions, by the key a shape uses to skip them. Each is one operation (or,
// for "subject", no operation at all) in the generated spec.
var allPositions = []string{
	"subject",               // the component schema Subject itself
	"holder",                // Subject as a $ref property, array items and additionalProperties value
	"holder-inline",         // the subject schema inlined as a property, array items and additionalProperties value
	"body-ref",              // request and response body: $ref Subject
	"body-inline",           // request and response body: the subject schema inline
	"body-component",        // components/requestBodies and components/responses wrapping $ref Subject
	"body-component-inline", // components/requestBodies and components/responses wrapping the inline subject
	"body-default",          // a `default:` response, so the strict envelope is a struct with Body and StatusCode
	"body-headers",          // a response with a header, so the strict envelope is a struct with Body and Headers
}

// newestVersion is the schema-merging-behavior version under development. The
// tests of every version before it guard against regressions, and say so.
const newestVersion = "v3"

// allVersions are the versions a shape can list.
var allVersions = []string{"v1", "v2", "v3"}

type shape struct {
	// Name is the file name without .yaml; it names the output directory.
	Name string `yaml:"-"`
	// Version is the schema-merging-behavior version being written.
	Version string `yaml:"-"`
	// Versions lists the schema-merging-behavior versions the shape runs
	// under; v2 when empty.
	Versions []string `yaml:"versions"`
	// Description is copied into the package doc comment.
	Description string `yaml:"description"`
	// OpenAPI is the spec version; 3.0.3 when empty.
	OpenAPI string `yaml:"openapi"`
	// Subject is the schema under test.
	Subject yaml.Node `yaml:"subject"`
	// Components are supporting component schemas.
	Components yaml.Node `yaml:"components"`
	// Config is merged into the generated oapi-codegen config.
	Config yaml.Node `yaml:"config"`
	// Samples are JSON documents that are valid instances of Subject. Each
	// must round-trip unchanged: include every required property and no
	// value the generated types cannot represent.
	Samples []string `yaml:"samples"`
	// Skip maps a position to the reason it is left out of this shape.
	Skip map[string]string `yaml:"skip"`
	// Broken maps a version older than the newest to the reason its
	// generated code can't round-trip the samples. The code is still
	// generated and compiled, so its output stays pinned, but the round
	// trips are skipped.
	Broken map[string]string `yaml:"broken"`
}

func (s shape) Has(position string) bool {
	_, skipped := s.Skip[position]
	return !skipped
}

// HasHolder reports whether either holder position is generated.
func (s shape) HasHolder() bool { return s.Has("holder") || s.Has("holder-inline") }

// HasOperations reports whether the spec has any operation, and so a client
// and a strict server.
func (s shape) HasOperations() bool {
	for _, p := range allPositions {
		if p != "subject" && s.Has(p) {
			return true
		}
	}
	return false
}

// HasPerSample reports whether any position runs one subtest per sample.
func (s shape) HasPerSample() bool {
	for _, p := range allPositions {
		if p != "holder" && p != "holder-inline" && s.Has(p) {
			return true
		}
	}
	return false
}

func (s shape) Package() string {
	return "matrix" + s.Version + strings.ReplaceAll(s.Name, "_", "")
}

// GuardsRegressions reports whether the version being written is older than
// the newest, so that its tests must not be changed to accept a regression.
func (s shape) GuardsRegressions() bool {
	return s.Version != newestVersion
}

// BrokenReason returns why the version being written can't round-trip the
// samples, or "".
func (s shape) BrokenReason() string {
	return s.Broken[s.Version]
}

func (s shape) Skips() []string {
	var out []string
	for _, p := range allPositions {
		if reason, ok := s.Skip[p]; ok {
			out = append(out, fmt.Sprintf("%s: %s", p, reason))
		}
	}
	return out
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "specgen:", err)
		os.Exit(1)
	}
}

func run() error {
	files, err := filepath.Glob(filepath.Join("shapes", "*.yaml"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no shapes found in %s; run from the matrix directory", filepath.Join(mustGetwd(), "shapes"))
	}
	sort.Strings(files)
	for _, f := range files {
		s, err := loadShape(f)
		if err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		for _, version := range s.Versions {
			s.Version = version
			if err := writeShape(s); err != nil {
				return fmt.Errorf("%s (%s): %w", f, version, err)
			}
		}
	}
	return nil
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func loadShape(path string) (shape, error) {
	var s shape
	raw, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&s); err != nil {
		return s, err
	}
	s.Name = strings.TrimSuffix(filepath.Base(path), ".yaml")
	if s.OpenAPI == "" {
		s.OpenAPI = "3.0.3"
	}
	if s.Subject.Kind == 0 {
		return s, fmt.Errorf("no subject")
	}
	if len(s.Samples) == 0 {
		return s, fmt.Errorf("no samples")
	}
	for i, sample := range s.Samples {
		if !json.Valid([]byte(sample)) {
			return s, fmt.Errorf("sample %d is not valid JSON: %s", i, sample)
		}
		s.Samples[i] = strings.TrimSpace(sample)
	}
	if len(s.Versions) == 0 {
		s.Versions = []string{"v2"}
	}
	for _, v := range s.Versions {
		if !slices.Contains(allVersions, v) {
			return s, fmt.Errorf("versions: unknown version %q", v)
		}
	}
	for v := range s.Broken {
		if !slices.Contains(s.Versions, v) || v == newestVersion {
			return s, fmt.Errorf("broken: %q must be a version this shape runs under, other than the newest", v)
		}
	}
	for p := range s.Skip {
		if !slices.Contains(allPositions, p) {
			return s, fmt.Errorf("skip: unknown position %q", p)
		}
	}
	return s, nil
}

func writeShape(s shape) error {
	dir := filepath.Join(s.Version, s.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	spec, err := buildSpec(s)
	if err != nil {
		return err
	}
	config, err := buildConfig(s)
	if err != nil {
		return err
	}
	harness, err := render(harnessTemplate, s, true)
	if err != nil {
		return fmt.Errorf("harness: %w", err)
	}
	doc, err := render(docTemplate, s, true)
	if err != nil {
		return fmt.Errorf("doc.go: %w", err)
	}
	outputs := map[string][]byte{
		"spec.yaml":          spec,
		"config.yaml":        config,
		"doc.go":             doc,
		"matrix_gen_test.go": harness,
	}
	for name, content := range outputs {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ---- YAML construction helpers ----

func str(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
}

func boolean(v bool) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprint(v)}
}

// mapping builds an ordered mapping node from alternating keys and values.
// Values may be *yaml.Node, string, or bool.
func mapping(kv ...any) *yaml.Node {
	n := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for i := 0; i < len(kv); i += 2 {
		n.Content = append(n.Content, str(kv[i].(string)), node(kv[i+1]))
	}
	return n
}

func node(v any) *yaml.Node {
	switch v := v.(type) {
	case *yaml.Node:
		return v
	case string:
		return str(v)
	case bool:
		return boolean(v)
	default:
		panic(fmt.Sprintf("unsupported yaml value %T", v))
	}
}

// clone deep-copies a node so each inline placement is independent.
func clone(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	c := *n
	c.Content = nil
	for _, child := range n.Content {
		c.Content = append(c.Content, clone(child))
	}
	c.Line, c.Column = 0, 0
	return &c
}

func ref(name string) *yaml.Node {
	return mapping("$ref", "#/components/schemas/"+name)
}

func jsonContent(schema *yaml.Node) *yaml.Node {
	return mapping("application/json", mapping("schema", schema))
}

func requestBody(schema *yaml.Node) *yaml.Node {
	return mapping("required", true, "content", jsonContent(schema))
}

func okResponse(schema *yaml.Node) *yaml.Node {
	return mapping("description", "ok", "content", jsonContent(schema))
}

func operation(id string, body, responses *yaml.Node) *yaml.Node {
	return mapping("post", mapping(
		"operationId", id,
		"requestBody", body,
		"responses", responses,
	))
}

// holderSchema places the subject (a $ref, or the schema itself) as a
// property, as array items and as an additionalProperties value.
func holderSchema(subject func() *yaml.Node) *yaml.Node {
	return mapping(
		"type", "object",
		"properties", mapping(
			"one", subject(),
			"list", mapping("type", "array", "items", subject()),
			"map", mapping("type", "object", "additionalProperties", subject()),
		),
	)
}

func buildSpec(s shape) ([]byte, error) {
	subjectRef := func() *yaml.Node { return ref("Subject") }
	subjectInline := func() *yaml.Node { return clone(&s.Subject) }

	paths := mapping()
	addPath := func(path string, op *yaml.Node) {
		paths.Content = append(paths.Content, str(path), op)
	}
	if s.Has("holder") {
		addPath("/holder", operation("holder",
			requestBody(ref("Holder")),
			mapping("200", okResponse(ref("Holder")))))
	}
	if s.Has("holder-inline") {
		addPath("/holder-inline", operation("holderInline",
			requestBody(ref("InlineHolder")),
			mapping("200", okResponse(ref("InlineHolder")))))
	}
	if s.Has("body-ref") {
		addPath("/body/ref", operation("bodyRef",
			requestBody(subjectRef()),
			mapping("200", okResponse(subjectRef()))))
	}
	if s.Has("body-inline") {
		addPath("/body/inline", operation("bodyInline",
			requestBody(subjectInline()),
			mapping("200", okResponse(subjectInline()))))
	}
	if s.Has("body-component") {
		addPath("/body/component", operation("bodyComponent",
			mapping("$ref", "#/components/requestBodies/SubjectBody"),
			mapping("200", mapping("$ref", "#/components/responses/SubjectResponse"))))
	}
	if s.Has("body-component-inline") {
		addPath("/body/component-inline", operation("bodyComponentInline",
			mapping("$ref", "#/components/requestBodies/InlineSubjectBody"),
			mapping("200", mapping("$ref", "#/components/responses/InlineSubjectResponse"))))
	}
	if s.Has("body-default") {
		addPath("/body/default", operation("bodyDefault",
			requestBody(subjectRef()),
			mapping("default", okResponse(subjectRef()))))
	}
	if s.Has("body-headers") {
		addPath("/body/headers", operation("bodyHeaders",
			requestBody(subjectRef()),
			mapping("200", mapping(
				"description", "ok",
				"headers", mapping("X-Shape", mapping("required", true, "schema", mapping("type", "string"))),
				"content", jsonContent(subjectRef())))))
	}

	schemas := mapping("Subject", clone(&s.Subject))
	if s.Has("holder") {
		schemas.Content = append(schemas.Content, str("Holder"), holderSchema(subjectRef))
	}
	if s.Has("holder-inline") {
		schemas.Content = append(schemas.Content, str("InlineHolder"), holderSchema(subjectInline))
	}
	if s.Components.Kind == yaml.MappingNode {
		for i := 0; i < len(s.Components.Content); i += 2 {
			schemas.Content = append(schemas.Content, clone(s.Components.Content[i]), clone(s.Components.Content[i+1]))
		}
	}

	components := mapping("schemas", schemas)
	requestBodies := mapping()
	responses := mapping()
	if s.Has("body-component") {
		requestBodies.Content = append(requestBodies.Content, str("SubjectBody"), requestBody(subjectRef()))
		responses.Content = append(responses.Content, str("SubjectResponse"), okResponse(subjectRef()))
	}
	if s.Has("body-component-inline") {
		requestBodies.Content = append(requestBodies.Content, str("InlineSubjectBody"), requestBody(subjectInline()))
		responses.Content = append(responses.Content, str("InlineSubjectResponse"), okResponse(subjectInline()))
	}
	if len(requestBodies.Content) > 0 {
		components.Content = append(components.Content, str("requestBodies"), requestBodies, str("responses"), responses)
	}

	doc := mapping(
		"openapi", s.OpenAPI,
		"info", mapping("title", s.Name, "version", "1.0.0"),
		"paths", paths,
		"components", components,
	)
	var buf bytes.Buffer
	buf.WriteString("# Code generated by specgen from ../../shapes/" + s.Name + ".yaml. DO NOT EDIT.\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildConfig(s shape) ([]byte, error) {
	cfg := mapping(
		"package", s.Package(),
		"output", s.Name+".gen.go",
		"generate", mapping(
			"models", true,
			"client", true,
			"std-http-server", true,
			"strict-server", true,
		),
	)
	if s.Config.Kind == yaml.MappingNode {
		for i := 0; i < len(s.Config.Content); i += 2 {
			if s.Config.Content[i].Value == "compatibility" {
				return nil, fmt.Errorf("config: compatibility is set per version; use versions")
			}
			cfg.Content = append(cfg.Content, clone(s.Config.Content[i]), clone(s.Config.Content[i+1]))
		}
	}
	cfg.Content = append(cfg.Content, str("compatibility"), mapping("schema-merging-behavior", s.Version))
	var buf bytes.Buffer
	buf.WriteString("# yaml-language-server: $schema=../../../../../../configuration-schema.json\n")
	buf.WriteString("# Code generated by specgen from ../../shapes/" + s.Name + ".yaml. DO NOT EDIT.\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(cfg); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func render(tmpl *template.Template, s shape, gofmt bool) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, s); err != nil {
		return nil, err
	}
	if !gofmt {
		return buf.Bytes(), nil
	}
	out, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%w\n%s", err, buf.String())
	}
	return out, nil
}

var funcs = template.FuncMap{
	"quote": func(s string) string { return fmt.Sprintf("%q", s) },
	"comment": func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			return ""
		}
		return "// " + strings.ReplaceAll(s, "\n", "\n// ")
	},
}

// docTemplate spells the directive as a template string so that go generate,
// which scans source lines, does not run it from this directory.
var docTemplate = template.Must(template.New("doc").Funcs(funcs).Parse(`// Code generated by specgen from ../../shapes/{{.Name}}.yaml. DO NOT EDIT.

// Package {{.Package}} is the {{.Name}} cell of the aggregate test matrix,
// generated with schema-merging-behavior {{.Version}}.
//
{{comment .Description}}
{{- if .GuardsRegressions}}
//
// The tests here record how {{.Version}} handles this shape. A bug fix may change
// the generated code and the tests, but the tests must not be changed to accept
// a regression: code that generated and worked before must keep doing so. A
// commit that changes them must say why.
{{- end}}
{{- if .BrokenReason}}
//
// The round trips are skipped: {{.BrokenReason}}
{{- end}}
{{- if .Skips}}
//
// Positions left out:
{{- range .Skips}}
//   - {{.}}
{{- end}}
{{- end}}
package {{.Package}}

{{"//go:generate"}} go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
`))

var harnessTemplate = template.Must(template.New("harness").Funcs(funcs).Parse(`// Code generated by specgen from ../../shapes/{{.Name}}.yaml. DO NOT EDIT.

package {{.Package}}
{{if .GuardsRegressions}}
// These tests record how schema-merging-behavior {{.Version}} handles this
// shape. Don't change them, or the shape they are generated from, to accept a
// regression; a commit that changes them must say why.
{{end}}
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// samples are valid instances of Subject that must survive every round trip
// unchanged.
var samples = []string{
{{- range .Samples}}
	{{quote .}},
{{- end}}
}

{{if .HasHolder -}}
// holderSample places the samples at every position of a holder schema.
func holderSample(t *testing.T) string {
	t.Helper()
	var list []json.RawMessage
	m := map[string]json.RawMessage{}
	for i, s := range samples {
		list = append(list, json.RawMessage(s))
		m[fmt.Sprintf("k%d", i)] = json.RawMessage(s)
	}
	b, err := json.Marshal(map[string]any{
		"one":  json.RawMessage(samples[0]),
		"list": list,
		"map":  m,
	})
	require.NoError(t, err)
	return string(b)
}
{{end -}}

{{if .HasOperations -}}
// echo re-types v through its JSON encoding, the way a handler that receives
// one generated type and must return another would.
func echo[T any](v any) (T, error) {
	var out T
	b, err := json.Marshal(v)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(b, &out)
	return out, err
}

type server struct{}

{{if .Has "holder" -}}
func (server) Holder(_ context.Context, req HolderRequestObject) (HolderResponseObject, error) {
	return echo[Holder200JSONResponse](req.Body)
}
{{end -}}
{{if .Has "holder-inline" -}}
func (server) HolderInline(_ context.Context, req HolderInlineRequestObject) (HolderInlineResponseObject, error) {
	return echo[HolderInline200JSONResponse](req.Body)
}
{{end -}}
{{if .Has "body-ref" -}}
func (server) BodyRef(_ context.Context, req BodyRefRequestObject) (BodyRefResponseObject, error) {
	return echo[BodyRef200JSONResponse](req.Body)
}
{{end -}}
{{if .Has "body-inline" -}}
func (server) BodyInline(_ context.Context, req BodyInlineRequestObject) (BodyInlineResponseObject, error) {
	return echo[BodyInline200JSONResponse](req.Body)
}
{{end -}}
{{if .Has "body-component" -}}
func (server) BodyComponent(_ context.Context, req BodyComponentRequestObject) (BodyComponentResponseObject, error) {
	body, err := echo[SubjectResponseJSONResponse](req.Body)
	return BodyComponent200JSONResponse{body}, err
}
{{end -}}
{{if .Has "body-component-inline" -}}
func (server) BodyComponentInline(_ context.Context, req BodyComponentInlineRequestObject) (BodyComponentInlineResponseObject, error) {
	body, err := echo[InlineSubjectResponseJSONResponse](req.Body)
	return BodyComponentInline200JSONResponse{body}, err
}
{{end -}}
{{if .Has "body-default" -}}
func (server) BodyDefault(_ context.Context, req BodyDefaultRequestObject) (BodyDefaultResponseObject, error) {
	body, err := echo[Subject](req.Body)
	return BodyDefaultdefaultJSONResponse{Body: body, StatusCode: http.StatusOK}, err
}
{{end -}}
{{if .Has "body-headers" -}}
func (server) BodyHeaders(_ context.Context, req BodyHeadersRequestObject) (BodyHeadersResponseObject, error) {
	body, err := echo[Subject](req.Body)
	return BodyHeaders200JSONResponse{Body: body, Headers: BodyHeaders200ResponseHeaders{XShape: "{{.Name}}"}}, err
}
{{end}}

func newClient(t *testing.T) *ClientWithResponses {
	t.Helper()
	srv := httptest.NewServer(Handler(NewStrictHandler(server{}, nil)))
	t.Cleanup(srv.Close)
	c, err := NewClientWithResponses(srv.URL)
	require.NoError(t, err)
	return c
}

// check asserts that both the bytes on the wire and the client's parsed
// response re-encode to want.
func check(t *testing.T, want string, status int, wire []byte, parsed any) {
	t.Helper()
	require.Equal(t, http.StatusOK, status, "body: %s", wire)
	assert.JSONEq(t, want, string(wire), "response bytes")
	got, err := json.Marshal(parsed)
	require.NoError(t, err)
	assert.JSONEq(t, want, string(got), "parsed response")
}
{{end -}}

{{if .HasPerSample -}}
// sampleName keeps subtest names short and readable.
func sampleName(i int, s string) string {
	if len(s) > 40 {
		s = s[:40]
	}
	return fmt.Sprintf("%d_%s", i, strings.NewReplacer(" ", "", "/", "_").Replace(s))
}
{{end}}

{{if .Has "subject" -}}
func TestMatrixSubject(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	for i, sample := range samples {
		t.Run(sampleName(i, sample), func(t *testing.T) {
			var v Subject
			require.NoError(t, json.Unmarshal([]byte(sample), &v))
			got, err := json.Marshal(v)
			require.NoError(t, err)
			assert.JSONEq(t, sample, string(got))
		})
	}
}
{{end -}}

{{if .Has "holder" -}}
func TestMatrixHolder(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	sample := holderSample(t)
	var v Holder
	require.NoError(t, json.Unmarshal([]byte(sample), &v))
	got, err := json.Marshal(v)
	require.NoError(t, err)
	assert.JSONEq(t, sample, string(got), "direct")

	resp, err := newClient(t).HolderWithResponse(context.Background(), v)
	require.NoError(t, err)
	check(t, sample, resp.StatusCode(), resp.Body, resp.JSON200)
}
{{end -}}

{{if .Has "holder-inline" -}}
func TestMatrixHolderInline(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	sample := holderSample(t)
	var v InlineHolder
	require.NoError(t, json.Unmarshal([]byte(sample), &v))
	got, err := json.Marshal(v)
	require.NoError(t, err)
	assert.JSONEq(t, sample, string(got), "direct")

	resp, err := newClient(t).HolderInlineWithResponse(context.Background(), v)
	require.NoError(t, err)
	check(t, sample, resp.StatusCode(), resp.Body, resp.JSON200)
}
{{end -}}

{{if .Has "body-ref" -}}
func TestMatrixBodyRef(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	c := newClient(t)
	for i, sample := range samples {
		t.Run(sampleName(i, sample), func(t *testing.T) {
			var body BodyRefJSONRequestBody
			require.NoError(t, json.Unmarshal([]byte(sample), &body))
			resp, err := c.BodyRefWithResponse(context.Background(), body)
			require.NoError(t, err)
			check(t, sample, resp.StatusCode(), resp.Body, resp.JSON200)
		})
	}
}
{{end -}}

{{if .Has "body-inline" -}}
func TestMatrixBodyInline(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	c := newClient(t)
	for i, sample := range samples {
		t.Run(sampleName(i, sample), func(t *testing.T) {
			var body BodyInlineJSONRequestBody
			require.NoError(t, json.Unmarshal([]byte(sample), &body))
			resp, err := c.BodyInlineWithResponse(context.Background(), body)
			require.NoError(t, err)
			check(t, sample, resp.StatusCode(), resp.Body, resp.JSON200)
		})
	}
}
{{end -}}

{{if .Has "body-component" -}}
func TestMatrixBodyComponent(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	c := newClient(t)
	for i, sample := range samples {
		t.Run(sampleName(i, sample), func(t *testing.T) {
			var body BodyComponentJSONRequestBody
			require.NoError(t, json.Unmarshal([]byte(sample), &body))
			resp, err := c.BodyComponentWithResponse(context.Background(), body)
			require.NoError(t, err)
			check(t, sample, resp.StatusCode(), resp.Body, resp.JSON200)
		})
	}
}
{{end -}}

{{if .Has "body-component-inline" -}}
func TestMatrixBodyComponentInline(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	c := newClient(t)
	for i, sample := range samples {
		t.Run(sampleName(i, sample), func(t *testing.T) {
			var body BodyComponentInlineJSONRequestBody
			require.NoError(t, json.Unmarshal([]byte(sample), &body))
			resp, err := c.BodyComponentInlineWithResponse(context.Background(), body)
			require.NoError(t, err)
			check(t, sample, resp.StatusCode(), resp.Body, resp.JSON200)
		})
	}
}
{{end -}}

{{if .Has "body-default" -}}
func TestMatrixBodyDefault(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	c := newClient(t)
	for i, sample := range samples {
		t.Run(sampleName(i, sample), func(t *testing.T) {
			var body BodyDefaultJSONRequestBody
			require.NoError(t, json.Unmarshal([]byte(sample), &body))
			resp, err := c.BodyDefaultWithResponse(context.Background(), body)
			require.NoError(t, err)
			check(t, sample, resp.StatusCode(), resp.Body, resp.JSONDefault)
		})
	}
}
{{end -}}

{{if .Has "body-headers" -}}
func TestMatrixBodyHeaders(t *testing.T) {
	{{if $.BrokenReason}}t.Skip({{quote $.BrokenReason}}){{end}}
	c := newClient(t)
	for i, sample := range samples {
		t.Run(sampleName(i, sample), func(t *testing.T) {
			var body BodyHeadersJSONRequestBody
			require.NoError(t, json.Unmarshal([]byte(sample), &body))
			resp, err := c.BodyHeadersWithResponse(context.Background(), body)
			require.NoError(t, err)
			assert.Equal(t, "{{.Name}}", resp.HTTPResponse.Header.Get("X-Shape"))
			check(t, sample, resp.StatusCode(), resp.Body, resp.JSON200)
		})
	}
}
{{end -}}
`))
