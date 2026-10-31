package codegen

import (
	"errors"
	"fmt"
	"go/token"
	"regexp"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
)

// ValidateSpec walks a loaded OpenAPI document and rejects values that cannot be
// represented in the generated Go source.
//
// oapi-codegen copies many spec strings verbatim into the code it emits:
// parameter names, property names, media types and enum values become Go string
// literals and struct-tag contents; the `x-go-name`/`x-go-type`/`x-go-type-name`
// extensions become Go identifiers and type expressions. A value containing a
// quote, a backtick or a control character (such as a newline) would produce Go
// that does not compile. Catching that here turns an obscure compiler error on
// generated code into a clear message that points at the offending part of the
// spec.
//
// It is meant to run after any overlay has been applied and after tag/operation
// filtering and component pruning, so only values that will actually be emitted
// are checked. A `$ref` with a fragment is not followed: it resolves to a
// component that is validated at its own definition, or to a type in another
// package whose contents are never copied into the generated output. A
// whole-document `$ref` (one without a fragment, such as `./other.yaml`) is
// followed, because code generation inlines what it points at (see descend).
func ValidateSpec(spec *openapi3.T) error {
	if spec == nil {
		return nil
	}
	v := &specValidator{}
	v.walkDocument(spec)
	return errors.Join(v.errs...)
}

type specValidator struct {
	errs []error
	// seen holds the resolved values of the whole-document $refs that have
	// been walked, so that documents referring to each other are walked once.
	seen map[any]bool
}

// descend reports whether the validator should walk into a $ref's resolved
// value, given the ref string and the (non-nil) value it resolved to.
//
// An inline object (no ref) is always walked. A ref with a fragment, local
// (`#/components/schemas/Foo`) or external (`other.yaml#/components/schemas/Foo`),
// is not: a local one resolves to a component that is validated at its own
// definition, and an external one becomes a type in another Go package whose
// contents are never copied into the generated output. A whole-document ref
// (`./other.yaml`, with no fragment) has no name that could be referred to, so
// code generation inlines its contents as if they were written in place
// (IsGoTypeReference is false for it); everything in that document reaches the
// generated source and has to be checked like the inline object it becomes.
// Each such document is walked once, so documents that refer to each other
// do not loop.
func (v *specValidator) descend(ref string, value any) bool {
	if ref == "" {
		return true
	}
	if !IsWholeDocumentReference(ref) {
		return false
	}
	if v.seen[value] {
		return false
	}
	if v.seen == nil {
		v.seen = make(map[any]bool)
	}
	v.seen[value] = true
	return true
}

func (v *specValidator) addf(format string, args ...any) {
	v.errs = append(v.errs, fmt.Errorf(format, args...))
}

// checkText validates a free-form string that is copied verbatim into the
// generated code (a name, media type, enum value, path, or tag content). Such a
// value ends up inside Go string literals, backtick-delimited struct tags and
// line comments, so it may not contain a double quote, a backtick, or a control
// character (newlines included). `what` describes the location for the message.
func (v *specValidator) checkText(value, what string) {
	for _, r := range value {
		if r == '"' || r == '`' || unicode.IsControl(r) {
			v.addf("%s may not contain %s: %q", what, describeRune(r), value)
			return
		}
	}
}

// checkNoControl validates a string that is copied into the generated code but
// may legitimately contain quotes — for example a media type, whose parameters
// use quoted-strings (`application/ld+json; profile="…"`). Such values are
// emitted through strconv.Quote at their string-literal sinks, which handles the
// quotes safely, so only control characters (newlines that would break out of a
// line comment) are rejected here.
func (v *specValidator) checkNoControl(value, what string) {
	for _, r := range value {
		if unicode.IsControl(r) {
			v.addf("%s may not contain %s: %q", what, describeRune(r), value)
			return
		}
	}
}

// checkDiscriminatorValue validates a discriminator mapping key, explicit or
// implicit. Besides the places checkText guards, the value is written between
// the quotes of an interpreted Go string literal (`case "<value>":`) and into
// the JSON object the From* helpers stamp, where a backslash starts an escape
// sequence: `case "Dog\Name":` does not compile. So a backslash is rejected
// too, as the union generators already do for the value an inline variant
// pins.
func (v *specValidator) checkDiscriminatorValue(value, what string) {
	for _, r := range value {
		if r == '"' || r == '`' || r == '\\' || unicode.IsControl(r) {
			v.addf("%s may not contain %s: %q", what, describeRune(r), value)
			return
		}
	}
}

// checkIdentifier validates a value that becomes a Go identifier (the `x-go-name`
// and `x-go-type-name` extensions). It must be a valid, non-keyword Go
// identifier, which by definition excludes quotes, backticks and control
// characters.
func (v *specValidator) checkIdentifier(value, what string) {
	if !token.IsIdentifier(value) {
		v.addf("%s must be a valid Go identifier: %q", what, value)
	}
}

// checkGoType validates a value that becomes a Go type expression (the `x-go-type`
// extension). It legitimately contains characters an identifier cannot (`.`,
// `[]`, `*`, `map[...]`, `struct{}`, …), so instead of an identifier check it
// forbids the characters that would let the value escape the type position and
// introduce a new declaration or statement: quotes, backticks, semicolons and
// control characters.
func (v *specValidator) checkGoType(value, what string) {
	for _, r := range value {
		if r == '"' || r == '`' || r == ';' || unicode.IsControl(r) {
			v.addf("%s must be a Go type expression and may not contain %s: %q", what, describeRune(r), value)
			return
		}
	}
}

// responseKeyRE matches the response keys OpenAPI allows: a status code, a
// range such as 2XX, or default.
var responseKeyRE = regexp.MustCompile(`^(default|[1-5]XX|[1-9][0-9][0-9])$`)

// checkResponseKey validates a key of an operation's responses map. The key is
// copied into the generated code as part of type names and, for the client, as
// the right-hand side of a status comparison (`rsp.StatusCode == 200`), so it
// has to be one of the forms OpenAPI defines and nothing else.
func (v *specValidator) checkResponseKey(key, what string) {
	if !responseKeyRE.MatchString(key) {
		v.addf("%s must be a three-digit status code, a range such as 2XX, or default: %q", what, key)
	}
}

// checkTagKey validates the key of an `x-oapi-codegen-extra-tags` entry. Struct
// tag keys are space-separated tokens terminated by ':', so they may not contain
// spaces, colons, quotes, backticks or control characters, and may not be empty.
func (v *specValidator) checkTagKey(key, what string) {
	if key == "" {
		v.addf("%s may not be empty", what)
		return
	}
	for _, r := range key {
		if r == '"' || r == '`' || r == ':' || r == ' ' || unicode.IsControl(r) {
			v.addf("%s may not contain %s: %q", what, describeRune(r), key)
			return
		}
	}
}

// checkGoImportName validates the alias field of an `x-go-type-import` entry.
// A Go import alias is either absent (empty), the blank identifier `_`, the
// dot `.` (dot-import), or any valid Go identifier. It is emitted unquoted
// (%s) into the import block, so a newline or other control character would
// close the block and allow arbitrary code injection.
func (v *specValidator) checkGoImportName(value, what string) {
	if value == "" || value == "." {
		return
	}
	v.checkIdentifier(value, what)
}

// checkGoImportPath validates the path field of an `x-go-type-import` entry.
// The path is emitted with %q (strconv.Quote), so control characters and
// quotes are escaped before they reach the output file. We still reject
// control characters here to catch obviously malformed inputs early.
func (v *specValidator) checkGoImportPath(value, what string) {
	v.checkNoControl(value, what)
}

// checkSecurity validates the scopes named in security requirements. The
// provider names are emitted only after identifier sanitization, but scopes are
// copied verbatim into a generated `[]string` literal (via toStringArray), so
// they are checked as text.
func (v *specValidator) checkSecurity(reqs openapi3.SecurityRequirements, where string) {
	for _, req := range reqs {
		for _, provider := range SortedMapKeys(req) {
			for _, scope := range req[provider] {
				v.checkText(scope, fmt.Sprintf("security scope for %q in %s", provider, where))
			}
		}
	}
}

// mappedExplicitly reports whether the discriminator's mapping names ref as
// one of its values, in which case the mapping key is used rather than the
// name derived from the ref. This mirrors how the union generators decide
// between an explicit and an implicit mapping.
func mappedExplicitly(d *openapi3.Discriminator, ref string) bool {
	for _, target := range d.Mapping {
		if target.Ref == ref {
			return true
		}
	}
	return false
}

func describeRune(r rune) string {
	switch r {
	case '"':
		return "a double quote"
	case '`':
		return "a backtick"
	case '\\':
		return "a backslash"
	case '\n':
		return "a newline"
	case '\r':
		return "a carriage return"
	case '\t':
		return "a tab"
	case ';':
		return "a semicolon"
	case ':':
		return "a colon"
	case ' ':
		return "a space"
	}
	if unicode.IsControl(r) {
		return fmt.Sprintf("a control character (%U)", r)
	}
	return fmt.Sprintf("%q", r)
}

// checkExtensions validates the code-affecting extensions on any spec object.
// Values are parsed with the same helpers code generation uses, and only checked
// when the parse succeeds — a malformed extension is ignored during generation,
// so it cannot inject anything and need not be rejected here.
func (v *specValidator) checkExtensions(ext map[string]any, where string) {
	if ext == nil {
		return
	}
	if raw, ok := ext[extGoName]; ok {
		if s, err := extParseGoFieldName(raw); err == nil {
			v.checkIdentifier(s, "x-go-name in "+where)
		}
	}
	if raw, ok := ext[extGoTypeName]; ok {
		if s, err := extTypeName(raw); err == nil {
			v.checkIdentifier(s, "x-go-type-name in "+where)
		}
	}
	if raw, ok := ext[extPropGoType]; ok {
		if s, err := extString(raw); err == nil {
			v.checkGoType(s, "x-go-type in "+where)
		}
	}
	if raw, ok := ext[extPropExtraTags]; ok {
		if tags, err := extExtraTags(raw); err == nil {
			for _, k := range SortedMapKeys(tags) {
				v.checkTagKey(k, "x-oapi-codegen-extra-tags key in "+where)
				v.checkText(tags[k], fmt.Sprintf("x-oapi-codegen-extra-tags value for %q in %s", k, where))
			}
		}
	}
	if raw, ok := ext[extPropGoImport]; ok {
		if importMap, ok := raw.(map[string]any); ok {
			for k, val := range importMap {
				if s, ok := val.(string); ok {
					switch {
					case strings.EqualFold(k, "name"):
						v.checkGoImportName(s, "x-go-type-import name in "+where)
					case strings.EqualFold(k, "path"):
						v.checkGoImportPath(s, "x-go-type-import path in "+where)
					}
				}
			}
		}
	}
}

func (v *specValidator) walkDocument(spec *openapi3.T) {
	v.checkSecurity(spec.Security, "the root security requirements")
	if spec.Paths != nil {
		paths := spec.Paths.Map()
		for _, path := range SortedMapKeys(paths) {
			where := fmt.Sprintf("path %q", path)
			v.checkText(path, "OpenAPI path")
			v.walkPathItem(paths[path], where)
		}
	}
	for _, name := range SortedMapKeys(spec.Webhooks) {
		where := fmt.Sprintf("webhook %q", name)
		v.checkText(name, "webhook name")
		v.walkPathItem(spec.Webhooks[name], where)
	}
	if spec.Components != nil {
		v.walkComponents(spec.Components)
	}
}

// walkComponents validates every component. Component names become Go
// identifiers only after sanitization, but they are also copied verbatim into
// single-line doc comments ("X defines model for <name>."), so a name may not
// contain a control character; quotes are fine there.
func (v *specValidator) walkComponents(c *openapi3.Components) {
	for _, name := range SortedMapKeys(c.Schemas) {
		v.checkNoControl(name, "schema name")
		v.walkSchemaRef(c.Schemas[name], fmt.Sprintf("schema %q", name))
	}
	for _, name := range SortedMapKeys(c.Parameters) {
		v.checkNoControl(name, "component parameter name")
		v.walkParameterRef(c.Parameters[name], fmt.Sprintf("component parameter %q", name))
	}
	for _, name := range SortedMapKeys(c.RequestBodies) {
		v.checkNoControl(name, "component request body name")
		v.walkRequestBodyRef(c.RequestBodies[name], fmt.Sprintf("component request body %q", name))
	}
	for _, name := range SortedMapKeys(c.Responses) {
		v.checkNoControl(name, "component response name")
		v.walkResponseRef(c.Responses[name], fmt.Sprintf("component response %q", name))
	}
	for _, name := range SortedMapKeys(c.Headers) {
		v.checkNoControl(name, "component header name")
		v.walkHeaderRef(c.Headers[name], fmt.Sprintf("component header %q", name))
	}
	for _, name := range SortedMapKeys(c.Callbacks) {
		v.checkNoControl(name, "component callback name")
		v.walkCallbackRef(c.Callbacks[name], fmt.Sprintf("component callback %q", name))
	}
	for _, name := range SortedSecuritySchemeKeys(c.SecuritySchemes) {
		v.checkNoControl(name, "security scheme name")
	}
}

func (v *specValidator) walkPathItem(item *openapi3.PathItem, where string) {
	if item == nil {
		return
	}
	for _, p := range item.Parameters {
		v.walkParameterRef(p, where)
	}
	for method, op := range item.Operations() {
		v.walkOperation(op, fmt.Sprintf("%s %s", method, where))
	}
}

func (v *specValidator) walkOperation(op *openapi3.Operation, where string) {
	if op == nil {
		return
	}
	v.checkExtensions(op.Extensions, where)
	// The operationId becomes a Go identifier only after normalization, but
	// the Echo templates also copy it verbatim (quoted) as the per-operation
	// middleware key, so it is checked like other quoted text.
	v.checkNoControl(op.OperationID, "operationId in "+where)
	if op.Security != nil {
		v.checkSecurity(*op.Security, where)
	}
	for _, p := range op.Parameters {
		v.walkParameterRef(p, where)
	}
	v.walkRequestBodyRef(op.RequestBody, where)
	if op.Responses != nil {
		responses := op.Responses.Map()
		for _, name := range SortedMapKeys(responses) {
			v.checkResponseKey(name, "response key in "+where)
			v.walkResponseRef(responses[name], fmt.Sprintf("response %q in %s", name, where))
		}
	}
	for _, name := range SortedMapKeys(op.Callbacks) {
		v.checkText(name, fmt.Sprintf("callback name in %s", where))
		v.walkCallbackRef(op.Callbacks[name], fmt.Sprintf("callback %q in %s", name, where))
	}
}

func (v *specValidator) walkParameterRef(ref *openapi3.ParameterRef, where string) {
	if ref == nil {
		return
	}
	v.checkExtensions(ref.Extensions, where)
	if ref.Value == nil || !v.descend(ref.Ref, ref.Value) {
		return
	}
	p := ref.Value
	loc := fmt.Sprintf("parameter %q in %s", p.Name, where)
	v.checkText(p.Name, "parameter name in "+where)
	v.checkText(p.Style, "style of "+loc)
	v.checkExtensions(p.Extensions, loc)
	v.walkContent(p.Content, loc)
	v.walkSchemaRef(p.Schema, loc)
}

func (v *specValidator) walkRequestBodyRef(ref *openapi3.RequestBodyRef, where string) {
	if ref == nil {
		return
	}
	v.checkExtensions(ref.Extensions, where)
	if ref.Value == nil || !v.descend(ref.Ref, ref.Value) {
		return
	}
	loc := "request body in " + where
	v.checkExtensions(ref.Value.Extensions, loc)
	v.walkContent(ref.Value.Content, loc)
}

func (v *specValidator) walkResponseRef(ref *openapi3.ResponseRef, where string) {
	if ref == nil {
		return
	}
	v.checkExtensions(ref.Extensions, where)
	if ref.Value == nil || !v.descend(ref.Ref, ref.Value) {
		return
	}
	r := ref.Value
	for _, name := range SortedMapKeys(r.Headers) {
		v.checkText(name, "response header name in "+where)
		v.walkHeaderRef(r.Headers[name], fmt.Sprintf("header %q in %s", name, where))
	}
	v.walkContent(r.Content, where)
	v.checkExtensions(r.Extensions, where)
}

func (v *specValidator) walkHeaderRef(ref *openapi3.HeaderRef, where string) {
	if ref == nil {
		return
	}
	v.checkExtensions(ref.Extensions, where)
	if ref.Value == nil || !v.descend(ref.Ref, ref.Value) {
		return
	}
	v.checkExtensions(ref.Value.Extensions, where)
	v.checkText(ref.Value.Style, "style of header in "+where)
	v.walkContent(ref.Value.Content, where)
	v.walkSchemaRef(ref.Value.Schema, where)
}

// ValidateServerURLs checks the parts of the document's servers that the
// `server-urls` generator copies into Go source as identifiers: the x-go-name
// extension, the names of the variables the URL uses (function parameters and
// type names), the `{placeholder}` names a URL uses without declaring (also
// parameters), and the enum values of the used variables, which are appended
// to the variable's type name to form constant names. None of them is
// sanitized on the way, so names have to be Go identifiers as written and enum
// values have to be made of identifier characters. A declared variable whose
// placeholder does not appear in the URL produces no code and is not checked.
// The function is separate from ValidateSpec because it only applies when
// server URLs are generated: specs with, say, a hyphenated variable name are
// fine otherwise.
func ValidateServerURLs(spec *openapi3.T) error {
	if spec == nil {
		return nil
	}
	v := &specValidator{}
	for i, server := range spec.Servers {
		if server == nil {
			continue
		}
		where := fmt.Sprintf("server %d (%s)", i, server.URL)
		v.checkExtensions(server.Extensions, where)
		used := ServerObjectDefinition{OAPISchema: server}.UsedVariables()
		for _, name := range SortedMapKeys(used) {
			v.checkIdentifier(name, "server variable name in "+where)
			variable := used[name]
			if variable == nil {
				continue
			}
			for _, value := range variable.Enum {
				// The constant is <TypeName><Value>, so the value may start
				// with a digit but must otherwise be identifier characters.
				if value == "" || !token.IsIdentifier("_"+value) {
					v.addf("enum value of server variable %q in %s must consist of letters, digits and underscores: %q", name, where, value)
				}
			}
		}
		for name := range urlPlaceholders(server.URL) {
			if _, declared := server.Variables[name]; !declared {
				v.checkIdentifier(name, "undeclared URL placeholder in "+where)
			}
		}
	}
	return errors.Join(v.errs...)
}

func (v *specValidator) walkCallbackRef(ref *openapi3.CallbackRef, where string) {
	if ref == nil {
		return
	}
	v.checkExtensions(ref.Extensions, where)
	if ref.Value == nil || !v.descend(ref.Ref, ref.Value) {
		return
	}
	items := ref.Value.Map()
	for _, expr := range SortedMapKeys(items) {
		v.walkPathItem(items[expr], where)
	}
}

func (v *specValidator) walkContent(content openapi3.Content, where string) {
	for _, mediaType := range SortedMapKeys(content) {
		v.checkNoControl(mediaType, "content type in "+where)
		media := content[mediaType]
		if media == nil {
			continue
		}
		loc := fmt.Sprintf("content type %q in %s", mediaType, where)
		v.checkExtensions(media.Extensions, loc)
		v.walkSchemaRef(media.Schema, loc)
	}
}

func (v *specValidator) walkSchemaRef(ref *openapi3.SchemaRef, where string) {
	if ref == nil {
		return
	}
	// Extensions on the ref node itself are honoured by code generation even for
	// a $ref — e.g. an `x-go-type` sibling next to a `$ref` overrides the
	// referenced type — so they must be checked here. Whether the target is
	// walked too is decided by descend: only a whole-document ref is, since
	// its contents are inlined into the generated code.
	v.checkExtensions(ref.Extensions, where)
	if ref.Value == nil {
		return
	}
	// An x-go-type next to a whole-document ref, or at the root of the
	// document it points at, replaces the document: generateGoSchema returns
	// the override before it reads any of the document's contents, so none
	// of them reach the generated code. Only the root's own extensions are
	// checked then. This is decided before descend marks the document as
	// walked: another ref to the same document without an override inlines
	// it, and has to find it unvisited. An inline schema with x-go-type has
	// always been walked in full, and still is.
	if IsWholeDocumentReference(ref.Ref) {
		if _, overridden := combinedSchemaExtensions(ref)[extPropGoType]; overridden {
			v.checkExtensions(ref.Value.Extensions, where)
			return
		}
	}
	if !v.descend(ref.Ref, ref.Value) {
		return
	}
	v.walkSchema(ref.Value, where)
}

// unionView is the discriminated union a schema describes once its allOf, if
// any, has been merged: every discriminator and every oneOf/anyOf variant
// found on the schema itself or on any member, through nested allOf and
// through $ref members, which the merge resolves. The merge concatenates the
// members' oneOf and anyOf lists and lets the one member that declares a
// discriminator govern the result, so a $ref variant from one member and a
// discriminator from another form a discriminated union that neither member
// shows on its own.
type unionView struct {
	discriminators []*openapi3.Discriminator
	variants       []*openapi3.SchemaRef
}

// collectUnionView gathers s's unionView. The allOf members are included
// only when the merge in effect combines them (see allOfMergesUnions). seen
// stops at a composition that refers back to itself.
func collectUnionView(s *openapi3.Schema, view *unionView, mergeAllOf bool, seen map[*openapi3.Schema]bool) {
	if s == nil || seen[s] {
		return
	}
	seen[s] = true
	if s.Discriminator != nil {
		view.discriminators = append(view.discriminators, s.Discriminator)
	}
	view.variants = append(view.variants, s.OneOf...)
	view.variants = append(view.variants, s.AnyOf...)
	if !mergeAllOf {
		return
	}
	for _, member := range s.AllOf {
		if member != nil {
			collectUnionView(member.Value, view, mergeAllOf, seen)
		}
	}
}

// allOfMergesUnions reports whether the schema-merging behavior of the
// current run merges the members of an allOf into one schema before
// generating it, so that the members' oneOf and anyOf lists and the
// discriminator one of them declares form a single union. The v1 merge
// (`old-merge-schemas: true`) generates each member on its own instead, and a
// member's union then has only the discriminator the member itself declares.
// Like the generators, this reads the options of the current run; outside a
// run, the default merge is assumed.
func allOfMergesUnions() bool {
	return !globalState.options.Compatibility.OldMergeSchemas
}

// checkImplicitMappingKeys validates the discriminator mapping keys the union
// generators derive rather than read from the spec. A $ref variant the mapping
// does not name is mapped implicitly, under the last segment of its ref (the
// schema's name), and that key is written into the generated code exactly
// like an explicit one. The name itself is otherwise allowed to contain quotes
// (see walkComponents), and a ref into another document names a schema this
// walk never sees, so the derived key is checked here. A null-typed branch is
// a nullability marker the union generators skip before mapping, so it never
// becomes a key.
//
// The variants and the discriminator are taken from the schema's unionView,
// so a union the allOf merge assembles from several members is checked too.
// A $ref member's own union is also validated where it is defined; here it
// is seen as part of the composition it is merged into.
func (v *specValidator) checkImplicitMappingKeys(s *openapi3.Schema, where string) {
	var view unionView
	collectUnionView(s, &view, allOfMergesUnions(), map[*openapi3.Schema]bool{})
	if len(view.discriminators) == 0 {
		return
	}
	for _, variant := range view.variants {
		if variant == nil || variant.Ref == "" || isNullTypeSchema(variant.Value) {
			continue
		}
		explicit := false
		for _, d := range view.discriminators {
			if mappedExplicitly(d, variant.Ref) {
				explicit = true
				break
			}
		}
		if explicit {
			continue
		}
		v.checkDiscriminatorValue(RefPathToObjName(variant.Ref), fmt.Sprintf("implicit discriminator mapping key for %s in %s", variant.Ref, where))
	}
}

func (v *specValidator) walkSchema(s *openapi3.Schema, where string) {
	if s == nil {
		return
	}
	v.checkExtensions(s.Extensions, where)
	v.checkText(s.Format, "format in "+where)
	if s.Type != nil {
		for _, t := range s.Type.Slice() {
			v.checkText(t, "type in "+where)
		}
	}
	// Enum values are emitted through strconv.Quote in the const block, so quotes
	// in a value are handled safely; only control characters are rejected.
	for _, enumValue := range s.Enum {
		if str, ok := enumValue.(string); ok {
			v.checkNoControl(str, "enum value in "+where)
		}
	}
	if s.Discriminator != nil {
		v.checkText(s.Discriminator.PropertyName, "discriminator propertyName in "+where)
		for _, key := range SortedMapKeys(s.Discriminator.Mapping) {
			v.checkDiscriminatorValue(key, "discriminator mapping key in "+where)
		}
	}
	v.checkImplicitMappingKeys(s, where)
	for _, name := range SortedMapKeys(s.Properties) {
		v.checkText(name, "property name in "+where)
		v.walkSchemaRef(s.Properties[name], fmt.Sprintf("property %q in %s", name, where))
	}
	v.walkSchemaRef(s.Items, "items in "+where)
	if s.AdditionalProperties.Schema != nil {
		v.walkSchemaRef(s.AdditionalProperties.Schema, "additionalProperties in "+where)
	}
	for i, sub := range s.AllOf {
		v.walkSchemaRef(sub, fmt.Sprintf("allOf[%d] in %s", i, where))
	}
	for i, sub := range s.AnyOf {
		v.walkSchemaRef(sub, fmt.Sprintf("anyOf[%d] in %s", i, where))
	}
	for i, sub := range s.OneOf {
		v.walkSchemaRef(sub, fmt.Sprintf("oneOf[%d] in %s", i, where))
	}
	v.walkSchemaRef(s.Not, "not in "+where)
}
