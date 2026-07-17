// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package codegen

import (
	"bufio"
	"bytes"
	"cmp"
	"fmt"
	"hash/fnv"
	"maps"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

type ParameterDefinition struct {
	ParamName string // The original json parameter name, eg param_name
	In        string // Where the parameter is defined - path, header, cookie, query
	Required  bool   // Is this a required parameter?
	Spec      *openapi3.Parameter
	Schema    Schema

	// Shared is true for a parameter declared at the path-item level, which
	// is inherited by every method on the path. Its helper types are declared
	// once for the path item rather than once per operation, so operation type
	// collection skips them (issue #2090).
	Shared bool
}

// TypeDef is here as an adapter after a large refactoring so that I don't
// have to update all the templates. It returns the type definition for a parameter,
// without the leading '*' for optional ones.
func (pd ParameterDefinition) TypeDef() string {
	typeDecl := pd.Schema.TypeDecl()
	return typeDecl
}

// RequiresNilCheck indicates whether the generated property should have a nil check performed on it before other checks.
// This should be used in templates when performing `nil` checks, but NOT when i.e. determining if there should be an optional pointer given to the type - in that case, use `HasOptionalPointer`
func (pd ParameterDefinition) RequiresNilCheck() bool {
	return pd.ZeroValueIsNil() || pd.HasOptionalPointer()
}

// ZeroValueIsNil is a helper function to determine if the given Go type used
// for this property has `nil` as its Go zero value. Slices (OpenAPI `array`)
// and maps (OpenAPI `object` with only `additionalProperties`, rendered as
// `map[K]V`) both satisfy this — templates use it to decide whether to emit a
// nil-check before reading the field.
func (pd ParameterDefinition) ZeroValueIsNil() bool {
	if pd.Schema.OAPISchema == nil {
		return false
	}

	if schemaPrimaryType(pd.Schema.OAPISchema.Type).Is("array") {
		return true
	}

	return strings.HasPrefix(pd.Schema.GoType, "map[")
}

// JsonTag generates the JSON annotation to map GoType to json type name. If Parameter
// Foo is marshaled to json as "foo", this will create the annotation
// 'json:"foo"'
// It also includes any additional struct tags from x-oapi-codegen-extra-tags
// at the parameter or schema level (parameter-level takes precedence).
func (pd *ParameterDefinition) JsonTag() string {
	fieldTags := make(map[string]string)

	if pd.Required {
		fieldTags["json"] = pd.ParamName
	} else {
		fieldTags["json"] = pd.ParamName + ",omitempty"
	}

	// Merge x-oapi-codegen-extra-tags from schema level first, then parameter level
	// so that parameter-level takes precedence.
	if pd.Spec != nil && pd.Spec.Schema != nil && pd.Spec.Schema.Value != nil {
		if extension, ok := pd.Spec.Schema.Value.Extensions[extPropExtraTags]; ok {
			if tags, err := extExtraTags(extension); err == nil {
				for k, v := range tags {
					fieldTags[k] = v
				}
			}
		}
	}
	if pd.Spec != nil {
		if extension, ok := pd.Spec.Extensions[extPropExtraTags]; ok {
			if tags, err := extExtraTags(extension); err == nil {
				for k, v := range tags {
					fieldTags[k] = v
				}
			}
		}
	}

	keys := SortedMapKeys(fieldTags)
	tags := make([]string, len(keys))
	for i, k := range keys {
		tags[i] = fmt.Sprintf(`%s:"%s"`, k, fieldTags[k])
	}
	return "`" + strings.Join(tags, " ") + "`"
}

func (pd *ParameterDefinition) IsJson() bool {
	p := pd.Spec
	if len(p.Content) == 1 {
		for k := range p.Content {
			if util.IsMediaTypeJson(k) {
				return true
			}
		}
	}
	return false
}

func (pd *ParameterDefinition) IsPassThrough() bool {
	p := pd.Spec
	if len(p.Content) > 1 {
		return true
	}
	if len(p.Content) == 1 {
		return !pd.IsJson()
	}
	return false
}

func (pd *ParameterDefinition) IsStyled() bool {
	p := pd.Spec
	return p.Schema != nil
}

func (pd *ParameterDefinition) Style() string {
	style := pd.Spec.Style
	if style == "" {
		in := pd.Spec.In
		switch in {
		case "path", "header":
			return "simple"
		case "query", "cookie":
			return "form"
		default:
			panic("unknown parameter format")
		}
	}
	return style
}

func (pd *ParameterDefinition) Explode() bool {
	if pd.Spec.Explode == nil {
		in := pd.Spec.In
		switch in {
		case "path", "header":
			return false
		case "query", "cookie":
			return true
		default:
			panic("unknown parameter format")
		}
	}
	return *pd.Spec.Explode
}

// SchemaType returns the first OpenAPI type string for this parameter's schema (e.g. "string", "integer"),
// or empty string if unavailable.
func (pd *ParameterDefinition) SchemaType() string {
	if pd.Spec.Schema != nil && pd.Spec.Schema.Value != nil && pd.Spec.Schema.Value.Type != nil {
		if s := pd.Spec.Schema.Value.Type.Slice(); len(s) > 0 {
			return s[0]
		}
	}
	return ""
}

// SchemaFormat returns the OpenAPI format string for this parameter's schema (e.g. "byte", "date-time"),
// or empty string if unavailable.
func (pd *ParameterDefinition) SchemaFormat() string {
	if pd.Spec.Schema != nil && pd.Spec.Schema.Value != nil {
		return pd.Spec.Schema.Value.Format
	}
	return ""
}

// SanitizedParamName returns the parameter name sanitized to be a valid Go
// identifier. This is needed for routers like net/http's ServeMux where path
// wildcards (e.g. {name}) must be valid Go identifiers. For the original
// OpenAPI parameter name (e.g. for error messages or JSON tags), use ParamName.
func (pd ParameterDefinition) SanitizedParamName() string {
	return SanitizeGoIdentifier(pd.ParamName)
}

func (pd ParameterDefinition) GoVariableName() string {
	name := LowercaseFirstCharacters(pd.GoName())
	if IsGoKeyword(name) {
		name = "p" + UppercaseFirstCharacter(name)
	}
	if unicode.IsNumber([]rune(name)[0]) {
		name = "n" + name
	}
	return name
}

func (pd ParameterDefinition) GoName() string {
	goName := pd.ParamName
	if extension, ok := pd.Spec.Extensions[extGoName]; ok {
		if extGoFieldName, err := extParseGoFieldName(extension); err == nil {
			goName = extGoFieldName
		}
	}
	return SchemaNameToTypeName(goName)
}

// Deprecated: Use HasOptionalPointer, as it is clearer what the intent is.
func (pd ParameterDefinition) IndirectOptional() bool {
	return !pd.Required && !pd.Schema.SkipOptionalPointer
}

// HasOptionalPointer indicates whether the generated property has an optional pointer associated with it.
// This takes into account the `x-go-type-skip-optional-pointer` extension, allowing a parameter definition to control whether the pointer should be skipped.
func (pd ParameterDefinition) HasOptionalPointer() bool {
	return !pd.Required && !pd.Schema.SkipOptionalPointer
}

type ParameterDefinitions []ParameterDefinition

func (p ParameterDefinitions) FindByName(name string) *ParameterDefinition {
	for _, param := range p {
		if param.ParamName == name {
			return &param
		}
	}
	return nil
}

// DescribeParameters walks the given parameters dictionary, and generates the above
// descriptors into a flat list. This makes it a lot easier to traverse the
// data in the template engine.
func DescribeParameters(params openapi3.Parameters, path []string) ([]ParameterDefinition, error) {
	outParams := make([]ParameterDefinition, 0, len(params))
	for _, paramOrRef := range params {
		param := paramOrRef.Value

		goType, err := paramToGoType(param, append(path, param.Name))
		if err != nil {
			return nil, fmt.Errorf("error generating type for param (%s): %s",
				param.Name, err)
		}

		pd := ParameterDefinition{
			ParamName: param.Name,
			In:        param.In,
			Required:  param.Required,
			Spec:      param,
			Schema:    goType,
		}

		// A parameter-level `x-go-type-skip-optional-pointer` overrides the
		// schema-level setting. `GenStructFromSchema` applies the same override
		// when rendering the params struct; without mirroring it here, the
		// client/server templates disagree with the struct definition and emit
		// a dereference (`*params.Field`) on a field declared without a pointer.
		if extension, ok := param.Extensions[extPropGoTypeSkipOptionalPointer]; ok {
			if skipOptionalPointer, err := extParsePropGoTypeSkipOptionalPointer(extension); err == nil {
				pd.Schema.SkipOptionalPointer = skipOptionalPointer
			}
		}

		// If this is a reference to a predefined type, simply use the reference
		// name as the type. $ref: "#/components/schemas/custom_type" becomes
		// "CustomType".
		if IsGoTypeReference(paramOrRef.Ref) {
			goType, err := RefPathToGoType(paramOrRef.Ref)
			if err != nil {
				return nil, fmt.Errorf("error dereferencing (%s) for param (%s): %s",
					paramOrRef.Ref, param.Name, err)
			}
			pd.Schema.GoType = goType
		}
		outParams = append(outParams, pd)
	}
	return outParams, nil
}

// paramNeedsHoisting reports whether a parameter's schema produces a named
// helper type — an anyOf/oneOf union member, an inline object, etc. — as
// opposed to a bare primitive. Only such parameters can cause the redeclaration
// collision fixed for issue #2090, so only they participate in collision
// detection. The check is exact: it describes the parameter and asks whether it
// hoisted anything, rather than second-guessing which schema shapes hoist.
func paramNeedsHoisting(paramRef *openapi3.ParameterRef) (bool, error) {
	if paramRef == nil || paramRef.Value == nil {
		return false, nil
	}
	described, err := DescribeParameters(openapi3.Parameters{paramRef}, nil)
	if err != nil {
		return false, err
	}
	for _, pd := range described {
		if len(pd.Schema.AdditionalTypes) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// sharedParamScope is one path item whose path-item-level parameters are shared
// by every method on it. hashKey is a string unique to the scope, hashed to
// disambiguate colliding names; source is a human-readable identifier for the
// scope used in generated doc comments.
type sharedParamScope struct {
	item    *openapi3.PathItem
	hashKey string
	source  string
}

// enumerateSharedParamScopes returns every path item in the spec that can
// declare path-item-level (shared) parameters: regular paths, webhooks, and
// callback path items. All three share the one global Go type namespace, so
// they must be considered together when detecting and resolving collisions.
func enumerateSharedParamScopes(swagger *openapi3.T) []sharedParamScope {
	var scopes []sharedParamScope

	if swagger.Paths != nil {
		for _, requestPath := range SortedMapKeys(swagger.Paths.Map()) {
			item := swagger.Paths.Value(requestPath)
			if item != nil {
				scopes = append(scopes, sharedParamScope{item: item, hashKey: requestPath, source: requestPath})
			}
		}
	}
	for _, webhookName := range SortedMapKeys(swagger.Webhooks) {
		if item := swagger.Webhooks[webhookName]; item != nil {
			scopes = append(scopes, sharedParamScope{item: item, hashKey: "webhook:" + webhookName, source: "webhook " + webhookName})
		}
	}
	if swagger.Paths != nil {
		for _, requestPath := range SortedMapKeys(swagger.Paths.Map()) {
			pathItem := swagger.Paths.Value(requestPath)
			if pathItem == nil {
				continue
			}
			for _, parentMethod := range SortedMapKeys(pathItem.Operations()) {
				parentOp := pathItem.Operations()[parentMethod]
				for _, cbName := range SortedMapKeys(parentOp.Callbacks) {
					cbRef := parentOp.Callbacks[cbName]
					if cbRef == nil || cbRef.Value == nil {
						continue
					}
					cb := cbRef.Value
					cbKeys := append([]string(nil), cb.Keys()...)
					slices.Sort(cbKeys)
					for _, urlExpr := range cbKeys {
						if item := cb.Value(urlExpr); item != nil {
							scopes = append(scopes, sharedParamScope{
								item:    item,
								hashKey: "callback:" + requestPath + ":" + parentMethod + ":" + cbName + ":" + urlExpr,
								source:  "callback " + cbName + " " + urlExpr,
							})
						}
					}
				}
			}
		}
	}
	return scopes
}

// resolveSharedParameters is the pre-pass for shared (path-item-level)
// parameter naming (issue #2090). It describes each scope's shared parameters
// once and returns them keyed by path item, ready for the operation loops to
// attach.
//
// A shared parameter's helper types are declared once for its path item, so
// two methods on the same path never collide. Names still collide *across*
// path items, though — the same `{id}` reused by sibling paths is the common
// case — so a helper-type name produced by more than one scope is disambiguated
// by prefixing every colliding scope's parameters with a short, stable hash of
// the scope (git-style: extended to the full hash only if two scopes' short
// hashes clash). A parameter that doesn't collide keeps its historical
// undecorated name, so existing generated code is unaffected.
func resolveSharedParameters(swagger *openapi3.T) (map[*openapi3.PathItem][]ParameterDefinition, error) {
	scopes := enumerateSharedParamScopes(swagger)

	// Count, per shared-parameter name, how many scopes hoist a helper type
	// under it. A name produced by two or more scopes collides.
	nameCounts := map[string]int{}
	for _, scope := range scopes {
		if len(scope.item.Operations()) == 0 {
			continue
		}
		for _, paramRef := range scope.item.Parameters {
			hoists, err := paramNeedsHoisting(paramRef)
			if err != nil {
				return nil, err
			}
			if hoists {
				nameCounts[paramRef.Value.Name]++
			}
		}
	}
	colliding := map[string]bool{}
	for name, n := range nameCounts {
		if n >= 2 {
			colliding[name] = true
		}
	}

	// Assign a disambiguating token to every scope that carries a colliding
	// parameter. Tokens are hashes of the scope key; a short prefix is used
	// unless two scopes' short prefixes clash, in which case both fall back to
	// the full hash.
	tokenScopeKeys := map[*openapi3.PathItem]string{}
	var needToken []string
	for _, scope := range scopes {
		if len(scope.item.Operations()) == 0 {
			continue
		}
		for _, paramRef := range scope.item.Parameters {
			if paramRef.Value != nil && colliding[paramRef.Value.Name] {
				tokenScopeKeys[scope.item] = scope.hashKey
				needToken = append(needToken, scope.hashKey)
				break
			}
		}
	}
	tokens := assignScopeTokens(needToken)

	result := map[*openapi3.PathItem][]ParameterDefinition{}
	for _, scope := range scopes {
		if _, done := result[scope.item]; done {
			continue
		}
		token := ""
		if key, ok := tokenScopeKeys[scope.item]; ok {
			token = tokens[key]
		}
		described, err := describeSharedParameters(scope.item.Parameters, scope.source, token, colliding)
		if err != nil {
			return nil, err
		}
		markShared(described)
		result[scope.item] = described
	}
	return result, nil
}

// assignScopeTokens maps each scope key to a stable identifier token derived
// from an FNV hash of the key. It uses a short prefix of the hash, extending
// any keys whose short prefixes collide to the full hash (git-style). The
// tokens are lowercase so they camel-case cleanly when threaded through type
// naming (e.g. "a1b2c3d" becomes the "A1b2c3d" prefix of "A1b2c3dId0").
func assignScopeTokens(keys []string) map[string]string {
	const shortLen = 7
	full := map[string]string{}
	short := map[string]string{}
	for _, key := range keys {
		h := fnv.New64a()
		_, _ = h.Write([]byte(key))
		sum := fmt.Sprintf("h%x", h.Sum64())
		full[key] = sum
		short[key] = sum[:min(shortLen+1, len(sum))]
	}
	shortCounts := map[string]int{}
	for _, key := range keys {
		shortCounts[short[key]]++
	}
	tokens := map[string]string{}
	for _, key := range keys {
		if shortCounts[short[key]] > 1 {
			tokens[key] = full[key]
		} else {
			tokens[key] = short[key]
		}
	}
	return tokens
}

// describeSharedParameters describes a scope's shared parameters once. A
// parameter whose name collides across scopes is prefixed with the scope's
// disambiguating token so its helper types are uniquely named; a parameter
// that does not collide keeps its historical undecorated name (issue #2090).
// For the disambiguated types, a doc comment is set explaining the otherwise
// opaque hash prefix and pointing back to the source path.
func describeSharedParameters(params openapi3.Parameters, source, token string, colliding map[string]bool) ([]ParameterDefinition, error) {
	out := make([]ParameterDefinition, 0, len(params))
	for _, paramRef := range params {
		collides := token != "" && paramRef.Value != nil && colliding[paramRef.Value.Name]
		var path []string
		if collides {
			path = []string{token}
		}
		described, err := DescribeParameters(openapi3.Parameters{paramRef}, path)
		if err != nil {
			return nil, err
		}
		if collides {
			for i := range described {
				for j := range described[i].Schema.AdditionalTypes {
					td := &described[i].Schema.AdditionalTypes[j]
					td.Comment = fmt.Sprintf(
						"// %s is a helper type for the shared %q parameter of %q, prefixed with a per-path hash to disambiguate it from the same-named parameter on another path.",
						td.TypeName, paramRef.Value.Name, source)
				}
			}
		}
		out = append(out, described...)
	}
	return out, nil
}

// markShared flags every parameter as shared at the path-item level, so its
// helper types are declared once for the path item rather than once per
// operation (issue #2090).
func markShared(params []ParameterDefinition) {
	for i := range params {
		params[i].Shared = true
	}
}

// sharedParameterTypeDefs returns the helper TypeDefinitions produced by
// path-item-level parameters. These are emitted once for the path item
// (attributed to its first operation) instead of once per operation.
func sharedParameterTypeDefs(sharedParams []ParameterDefinition) []TypeDefinition {
	var typeDefs []TypeDefinition
	for _, param := range sharedParams {
		typeDefs = append(typeDefs, param.Schema.AdditionalTypes...)
	}
	return typeDefs
}

type SecurityDefinition struct {
	ProviderName string
	Scopes       []string
}

func DescribeSecurityDefinition(securityRequirements openapi3.SecurityRequirements) []SecurityDefinition {
	outDefs := make([]SecurityDefinition, 0)

	for _, sr := range securityRequirements {
		if len(sr) == 0 {
			return nil
		}
		for _, k := range SortedMapKeys(sr) {
			v := sr[k]
			outDefs = append(outDefs, SecurityDefinition{ProviderName: k, Scopes: v})
		}
	}

	return outDefs
}

// filterOutUndefinedSecuritySchemes drops any SecurityDefinition whose ProviderName
// is not present in defined. A `security` requirement that references an
// unknown scheme would otherwise produce a constant declaration and middleware
// references against a context-key type that is never emitted (the type is
// only generated for entries in components/securitySchemes).
func filterOutUndefinedSecuritySchemes(defs []SecurityDefinition, defined map[string]struct{}) []SecurityDefinition {
	out := make([]SecurityDefinition, 0, len(defs))
	for _, d := range defs {
		if _, ok := defined[d.ProviderName]; ok {
			out = append(out, d)
		}
	}
	return out
}

// OperationDefinition describes an Operation
type OperationDefinition struct {
	// OperationId is the `operationId` field from the OpenAPI Specification, after going through a `nameNormalizer`, and will be used to generate function names
	OperationId string
	// SpecOperationId is the raw `operationId` value as it appears in the OpenAPI spec, before normalization to a Go identifier. Empty when the spec didn't supply one (in which case the codegen-generated ID is the only available identifier and is exposed via OperationId).
	SpecOperationId string

	PathParams          []ParameterDefinition // Parameters in the path, eg, /path/:param
	HeaderParams        []ParameterDefinition // Parameters in HTTP headers
	QueryParams         []ParameterDefinition // Parameters in the query, /path?param
	CookieParams        []ParameterDefinition // Parameters in cookies
	TypeDefinitions     []TypeDefinition      // These are all the types we need to define for this operation
	SecurityDefinitions []SecurityDefinition  // These are the security providers
	BodyRequired        bool
	Bodies              []RequestBodyDefinition // The list of bodies for which to generate handlers.
	Responses           []ResponseDefinition    // The list of responses that can be accepted by handlers.
	Summary             string                  // Summary string from Swagger, used to generate a comment
	Method              string                  // GET, POST, DELETE, etc.
	Path                string                  // The Swagger path for the operation, like /resource/{id}
	// SpecOrder is the source line on which this operation's path is
	// declared in the spec, used to register routes in the order the paths
	// appear in the spec rather than sorted (issue #1887). Zero when the
	// source location is unavailable (e.g. a programmatically-built spec),
	// in which case route registration falls back to the default order.
	SpecOrder int
	Spec      *openapi3.Operation
	IsAlias             bool   // True when this path is a $ref alias of another path item
	AliasTarget         string // When IsAlias is true, this is the OperationId of the canonical operation (for route registration to reference the correct wrapper)
	PathItemRef         string // The path item's $ref (if any); used to qualify externally-loaded schemas referenced from this operation's responses

	// IsWebhook is true when this OperationDefinition was sourced from
	// spec.Webhooks (OpenAPI 3.1+). Webhook operations have no path
	// template; the target URL is supplied per-call by the initiator.
	IsWebhook bool
	// WebhookName is the spec.Webhooks map key when IsWebhook is true.
	WebhookName string

	// IsCallback is true when this OperationDefinition was sourced from
	// a parent operation's `callbacks:` block (OpenAPI 3.0+). Callback
	// operations have no path template at codegen time; the target URL
	// is the runtime callback URL discovered via the spec's callback
	// expression (typically a field on the parent operation's request
	// body) and is supplied per-call by the initiator.
	IsCallback bool
	// CallbackName is the parent operation's `callbacks:` map key
	// (e.g. "treePlanted") when IsCallback is true.
	CallbackName string
}

// HandlerName returns the OperationId to use when referencing the server-side
// wrapper function. For alias operations this is the canonical operation's ID,
// since the alias doesn't generate its own wrapper.
func (o *OperationDefinition) HandlerName() string {
	if o.IsAlias {
		return o.AliasTarget
	}
	return o.OperationId
}

// MiddlewareKey returns the identifier to use as the key in per-operation
// middleware maps. The raw spec OperationId is preferred so map keys mirror
// the OpenAPI spec verbatim; falls back to the normalized OperationId when
// the spec didn't supply one.
func (o *OperationDefinition) MiddlewareKey() string {
	if o.SpecOperationId != "" {
		return o.SpecOperationId
	}
	return o.OperationId
}

// SourceName returns WebhookName when IsWebhook, CallbackName when
// IsCallback, or empty otherwise. Templates use this to label the
// emitted handler uniformly without branching on which kind of source
// the operation came from.
func (o OperationDefinition) SourceName() string {
	if o.IsWebhook {
		return o.WebhookName
	}
	if o.IsCallback {
		return o.CallbackName
	}
	return ""
}

// ReceiverTemplateData is the input to the per-framework webhook /
// callback receiver template. Prefix selects between "Webhook" and
// "Callback" (and the lowercase form for prose), so a single template
// per framework handles both kinds.
type ReceiverTemplateData struct {
	Prefix      string // "Webhook" or "Callback"
	PrefixLower string // lowercase form of Prefix, for prose
	Operations  []OperationDefinition
}

// NewReceiverTemplateData builds the template input for the given
// prefix ("Webhook" or "Callback") and operation list.
func NewReceiverTemplateData(prefix string, ops []OperationDefinition) ReceiverTemplateData {
	return ReceiverTemplateData{
		Prefix:      prefix,
		PrefixLower: strings.ToLower(prefix),
		Operations:  ops,
	}
}

// Params returns the list of all parameters except Path parameters. Path parameters
// are handled differently from the rest, since they're mandatory.
func (o *OperationDefinition) Params() []ParameterDefinition {
	result := append(o.QueryParams, o.HeaderParams...)
	result = append(result, o.CookieParams...)
	return result
}

// AllParams returns all parameters
func (o *OperationDefinition) AllParams() []ParameterDefinition {
	result := append(o.QueryParams, o.HeaderParams...)
	result = append(result, o.CookieParams...)
	result = append(result, o.PathParams...)
	return result
}

// If we have parameters other than path parameters, they're bundled into an
// object. Returns true if we have any of those. This is used from the template
// engine.
func (o *OperationDefinition) RequiresParamObject() bool {
	return len(o.Params()) > 0
}

// HasBody is called by the template engine to determine whether to generate body
// marshaling code on the client. This is true for all body types, whether
// we generate types for them.
func (o *OperationDefinition) HasBody() bool {
	return o.Spec.RequestBody != nil
}

// SummaryAsComment returns the Operations summary as a Godoc-style multi line comment
func (o *OperationDefinition) SummaryAsComment(prefix string) string {
	if o.Summary == "" {
		return ""
	}
	parts := strings.Split(normalizeWhitespace(o.Summary), "\n")
	for i, p := range parts {
		if i == 0 && prefix != "" {
			parts[i] = "// " + prefix + " " + p
		} else {
			parts[i] = "// " + p
		}
	}
	return strings.Join(parts, "\n")
}

// prepareDescriptionLines normalises description for inclusion in a Godoc comment, and:
//
//  1. Returns nil when the description is the same as the summary
//  1. Ensures a single-line description includes a trailing `.`, which prevents
//     gofmt from promoting this to an "old-style" header
func prepareDescriptionLines(summary, description string) []string {
	description = normalizeWhitespace(description)
	if description == "" {
		return nil
	}
	if description == normalizeWhitespace(summary) {
		return nil
	}
	lines := strings.Split(description, "\n")
	if len(lines) == 1 && !strings.ContainsAny(lines[0], ".,;:!?") {
		lines[0] += "."
	}
	return lines
}

// GenerateFunctionComment returns a full Godoc-style multi-line comment with:
// - the Summary, if present, as the first line of the comment
// - if not present, an indication of the HTTP call this corresponds with
// - the Description, if present
// - whether this function takes a body and a content type
//
// Takes originalFunctionName (the OperationId or the function name being generated for this Operation), a suffix (if necessary) and whether this is being generated for ClientInterface or ClientWithResponsesInterface
func (o OperationDefinition) GenerateFunctionComment(originalFunctionName string, functionSuffix string, isFunctionWithResponses bool) string {
	functionName := originalFunctionName + functionSuffix
	descriptionLines := prepareDescriptionLines(o.Summary, o.Spec.Description)

	var parts []string
	if summary := o.SummaryAsComment(functionName); summary != "" {
		parts = append(parts, strings.Split(summary, "\n")...)
		parts = append(parts, "//")
		if len(descriptionLines) > 0 {
			for _, line := range descriptionLines {
				parts = append(parts, "// "+line)
			}
			parts = append(parts, "//")
		}
		if o.HasBody() {
			if isFunctionWithResponses {
				parts = append(parts, "// Takes any type of body and a specified content type, and returns a wrapper object for the known response body format(s).")
				parts = append(parts, "//")
			} else {
				parts = append(parts, "// Takes any type of body and a specified content type.")
				parts = append(parts, "//")
			}
		} else {
			if isFunctionWithResponses {
				parts = append(parts, "// Returns a wrapper object for the known response body format(s).")
				parts = append(parts, "//")
			}
		}
		parts = append(parts, "// Corresponds with "+o.Method+" "+o.Path+" (the `"+o.OperationId+"` operationId).")
	} else {
		if o.HasBody() {
			parts = append(parts, "// "+functionName+" performs a "+o.Method+" "+o.Path+" (the `"+o.OperationId+"` operationId) request,")
			parts = append(parts, "// with any type of body and a specified content type.")
		} else {
			parts = append(parts, "// "+functionName+" performs a "+o.Method+" "+o.Path+" (the `"+o.OperationId+"` operationId) request.")
		}
		if len(descriptionLines) > 0 {
			parts = append(parts, "//")
			for _, line := range descriptionLines {
				parts = append(parts, "// "+line)
			}
		}
		if isFunctionWithResponses {
			parts = append(parts, "//")
			parts = append(parts, "// Returns a wrapper object for the known response body format(s).")
		}
	}

	// make sure that each line is sanitised
	for i, part := range parts {
		parts[i] = stripNewLines(part)
	}

	return strings.Join(parts, "\n")
}

// DeprecationComment returns a Go-style deprecation comment if the operation is deprecated, otherwise returns an empty string.
func (o *OperationDefinition) DeprecationComment() string {
	if o.Spec == nil || !o.Spec.Deprecated {
		return ""
	}
	reason := "this operation has been marked as deprecated upstream, but no `x-deprecated-reason` was set"
	if extension, ok := o.Spec.Extensions[extDeprecationReason]; ok {
		if r, err := extParseDeprecationReason(extension); err == nil {
			reason = r
		}
	}
	return DeprecationComment(reason)
}

// responseMediaTypeSuffix returns the media-type discriminator suffix (e.g.
// "ApplicationJSON") that must be appended to the Go type name generated for a
// single content entry of a response, or "" when the base name is used as-is.
//
// A suffix is required only when one response carries more than one
// JSON-compatible content type, and only the JSON entries receive one: non-JSON
// media types (XML, YAML, ...) never get a dedicated generated type, so they
// reuse the base name. Both the type declaration (GenerateTypesForResponses) and
// the client response-wrapper field types (GetResponseTypeDefinitions) must make
// this identical decision — otherwise the wrapper references per-content-type
// type names that were never declared (see issue #2389).
func responseMediaTypeSuffix(content openapi3.Content, mediaType string) string {
	if !util.IsMediaTypeJson(mediaType) {
		return ""
	}
	jsonCount := 0
	for mt := range content {
		if util.IsMediaTypeJson(mt) {
			jsonCount++
		}
	}
	if jsonCount <= 1 {
		return ""
	}
	return mediaTypeToCamelCase(mediaType)
}

// GetResponseTypeDefinitions produces a list of type definitions for a given Operation for the response
// types which we know how to parse. These will be turned into fields on a
// response object for automatic deserialization of responses in the generated
// Client code. See "client-with-responses.tmpl".
func (o *OperationDefinition) GetResponseTypeDefinitions() ([]ResponseTypeDefinition, error) {
	var tds []ResponseTypeDefinition

	if o.Spec == nil || o.Spec.Responses == nil {
		return tds, nil
	}

	sortedResponsesKeys := SortedMapKeys(o.Spec.Responses.Map())
	for _, responseName := range sortedResponsesKeys {
		responseRef := o.Spec.Responses.Value(responseName)

		// We can only generate a type if we have a value:
		if responseRef.Value != nil {
			sortedContentKeys := SortedMapKeys(responseRef.Value.Content)
			for _, contentTypeName := range sortedContentKeys {
				contentType := responseRef.Value.Content[contentTypeName]
				// We can only generate a type if we have a schema:
				if contentType.Schema != nil {
					var typeName, tag string
					switch {

					// HAL+JSON:
					case slices.Contains(contentTypesHalJSON, contentTypeName):
						typeName = fmt.Sprintf("HALJSON%s", nameNormalizer(responseName))
						tag = "HALJSON"
					case contentTypeName == "application/json":
						// if it's the standard application/json
						typeName = fmt.Sprintf("JSON%s", nameNormalizer(responseName))
						tag = "JSON"
					// Vendored JSON
					case slices.Contains(contentTypesJSON, contentTypeName) || util.IsMediaTypeJson(contentTypeName):
						baseTypeName := fmt.Sprintf("%s%s", nameNormalizer(contentTypeName), nameNormalizer(responseName))

						typeName = strings.ReplaceAll(baseTypeName, "Json", "JSON")
						tag = mediaTypeToCamelCase(contentTypeName)
					// YAML:
					case slices.Contains(contentTypesYAML, contentTypeName):
						typeName = fmt.Sprintf("YAML%s", nameNormalizer(responseName))
						tag = "YAML"
					// XML:
					case slices.Contains(contentTypesXML, contentTypeName):
						typeName = fmt.Sprintf("XML%s", nameNormalizer(responseName))
						tag = "XML"
					default:
						continue
					}

					// Use the same body-type name as the server-side
					// GenerateResponseDefinitions ("Body" suffixed so it
					// doesn't collide with the strict envelope's struct
					// wrapper) as the schema-path root. The canonical
					// declaration happens server-side; here we just point
					// RefType at the same name so the JSON<status> field
					// renders as a pointer to it.
					responseBodyTypeName := o.OperationId + responseName + tag + "ResponseBody"
					schemaPath := []string{responseBodyTypeName}
					responseSchema, err := GenerateGoSchema(contentType.Schema, schemaPath)
					if err != nil {
						return nil, fmt.Errorf("unable to determine Go type for %s.%s: %w", o.OperationId, contentTypeName, err)
					}

					// Hoist inline response-root schemas that need
					// method-emitting boilerplate (UnionElements /
					// AdditionalProperties). For external path items,
					// qualify with the imported package — see the
					// equivalent block in GenerateResponseDefinitions for
					// rationale.
					if !IsGoTypeReference(responseRef.Ref) && responseSchema.RefType == "" &&
						(len(responseSchema.UnionElements) != 0 || responseSchema.HasAdditionalProperties ||
							(globalState.options.OutputOptions.GenerateTypesForAnonymousSchemas && len(responseSchema.Properties) > 0)) {
						if externalPkg := externalPackageFor(o.PathItemRef); externalPkg != "" {
							responseSchema.RefType = fmt.Sprintf("%s.%s", externalPkg, responseBodyTypeName)
						} else {
							responseSchema.RefType = responseBodyTypeName
						}
					}

					td := ResponseTypeDefinition{
						TypeDefinition: TypeDefinition{
							TypeName: typeName,
							Schema:   responseSchema,
						},
						ResponseName:              responseName,
						ContentTypeName:           contentTypeName,
						AdditionalTypeDefinitions: responseSchema.GetAdditionalTypeDefs(),
					}
					// A component response only declares a Go type for its JSON
					// content (GenerateTypesForResponses skips non-JSON media).
					// Point the wrapper field at that declared component type for
					// JSON content only; non-JSON content keeps the type derived
					// from its own schema above, so it neither references an
					// undeclared base type (e.g. an XML-only component response)
					// nor silently decodes into the JSON type when the JSON and
					// non-JSON schemas differ.
					if IsGoTypeReference(responseRef.Ref) && util.IsMediaTypeJson(contentTypeName) {
						// Determine the imported package for an external response.
						// Either the operation came from an externally-ref'd path
						// item and the response ref is relative to that file
						// (issue #2308), or the response ref targets an external
						// file directly (issue #2422).
						externalPkg := ""
						if pkg := externalPackageFor(o.PathItemRef); pkg != "" && strings.HasPrefix(responseRef.Ref, "#") {
							externalPkg = pkg
						} else if pkg := externalPackageFor(responseRef.Ref); pkg != "" {
							externalPkg = pkg
						}

						if externalPkg != "" {
							// The client wrapper points at the imported package's
							// response *model* type. Resolve its name in the
							// external document's context rather than via
							// RefPathToGoType / resolvedNameForRefPath, which key
							// off the root spec; in particular honour x-go-name on
							// the response component the same way the external
							// package generated it.
							refParts := strings.Split(responseRef.Ref, "/")
							refType := SchemaNameToTypeName(refParts[len(refParts)-1])
							if ext, ok := responseRef.Value.Extensions[extGoName]; ok {
								if name, err := extTypeName(ext); err == nil {
									refType = name
								}
							}
							refType += responseMediaTypeSuffix(responseRef.Value.Content, contentTypeName)
							td.Schema.RefType = fmt.Sprintf("%s.%s", externalPkg, refType)
						} else {
							refType, err := RefPathToGoType(responseRef.Ref)
							if err != nil {
								return nil, fmt.Errorf("error dereferencing response Ref: %w", err)
							}
							if suffix := responseMediaTypeSuffix(responseRef.Value.Content, contentTypeName); suffix != "" {
								if resolved := resolvedNameForRefPath(responseRef.Ref, contentTypeName); resolved != "" {
									refType = resolved + suffix
								} else {
									refType += suffix
								}
							}
							td.Schema.RefType = refType
						}
					}
					tds = append(tds, td)
				}
			}
		}
	}
	return tds, nil
}

func (o *OperationDefinition) HasMaskedRequestContentTypes() bool {
	return slices.ContainsFunc(o.Bodies, func(body RequestBodyDefinition) bool {
		return !body.IsFixedContentType()
	})
}

// RequestBodyDefinition describes a request body
type RequestBodyDefinition struct {
	// Is this body required, or optional?
	Required bool

	// This is the schema describing this body
	Schema Schema

	// When we generate type names, we need a Tag for it, such as JSON, in
	// which case we will produce "JSONBody".
	NameTag string

	// This is the content type corresponding to the body, eg, application/json
	ContentType string

	// Whether this is the default body type. For an operation named OpFoo, we
	// will not add suffixes like OpFooJSONBody for this one.
	Default bool

	// Contains encoding options for formdata
	Encoding map[string]RequestBodyEncoding

	// Deprecated indicates the parent operation is deprecated, so this body
	// type alias should be marked deprecated too.
	Deprecated bool

	// DeprecationReason is propagated from the parent operation's x-deprecated-reason.
	DeprecationReason string
}

// GenerateFunctionComment returns a full Godoc-style multi-line comment with:
// - the Summary, if present, as the first line of the comment
// - if not present, an indication of the HTTP call this corresponds with
// - whether this function takes a body and a content type
//
// Takes originalFunctionName (the OperationId or the function name being generated for this Operation), a suffix (if necessary) and whether this is being generated for ClientInterface or ClientWithResponsesInterface
func (r RequestBodyDefinition) GenerateFunctionComment(originalFunctionName string, parent OperationDefinition, functionSuffix string, isFunctionWithResponses bool) string {
	functionName := originalFunctionName + functionSuffix
	descriptionLines := prepareDescriptionLines(parent.Summary, parent.Spec.Description)

	var parts []string
	if summary := parent.SummaryAsComment(functionName); summary != "" {
		parts = append(parts, strings.Split(summary, "\n")...)
		parts = append(parts, "//")
		if len(descriptionLines) > 0 {
			for _, line := range descriptionLines {
				parts = append(parts, "// "+line)
			}
			parts = append(parts, "//")
		}
		if isFunctionWithResponses {
			parts = append(parts, "// Takes a body of the `"+r.ContentType+"` content type, and returns a wrapper object for the known response body format(s).")
		} else {
			parts = append(parts, "// Takes a body of the `"+r.ContentType+"` content type.")
		}
		parts = append(parts, "//")
		parts = append(parts, "// Corresponds with "+parent.Method+" "+parent.Path+" (the `"+parent.OperationId+"` operationId).")
	} else {
		parts = append(parts, "// "+functionName+" performs a "+parent.Method+" "+parent.Path+" (the `"+parent.OperationId+"` operationId) request.")
		if isFunctionWithResponses {
			parts = append(parts, "// Takes a body of the `"+r.ContentType+"` content type, and returns a wrapper object for the known response body format(s).")
		} else {
			parts = append(parts, "// Takes a body of the `"+r.ContentType+"` content type.")
		}
		if len(descriptionLines) > 0 {
			parts = append(parts, "//")
			for _, line := range descriptionLines {
				parts = append(parts, "// "+line)
			}
		}
	}

	// make sure that each line is sanitised
	for i, part := range parts {
		parts[i] = stripNewLines(part)
	}

	return strings.Join(parts, "\n")
}

// TypeDef returns the Go type definition for a request body
func (r RequestBodyDefinition) TypeDef(opID string) *TypeDefinition {
	return &TypeDefinition{
		TypeName:          fmt.Sprintf("%s%sRequestBody", opID, r.NameTag),
		Schema:            r.Schema,
		Deprecated:        r.Deprecated,
		DeprecationReason: r.DeprecationReason,
	}
}

// CustomType returns whether the body is a custom inline type, or pre-defined. This is
// poorly named, but it's here for compatibility reasons post-refactoring
// TODO: clean up the templates code, it can be simpler.
func (r RequestBodyDefinition) CustomType() bool {
	return r.Schema.RefType == ""
}

// When we're generating multiple functions which relate to request bodies,
// this generates the suffix. Such as Operation DoFoo would be suffixed with
// DoFooWithXMLBody.
func (r RequestBodyDefinition) Suffix() string {
	// The default response is never suffixed.
	if r.Default {
		return ""
	}
	return "With" + r.NameTag + "Body"
}

// IsSupportedByClient returns true if we support this content type for client. Otherwise only generic method will ge generated
func (r RequestBodyDefinition) IsSupportedByClient() bool {
	return r.IsJSON() || r.NameTag == "Formdata" || r.NameTag == "Text"
}

// IsJSON returns whether this is a JSON media type, for instance:
// - application/json
// - application/vnd.api+json
// - application/*+json
func (r RequestBodyDefinition) IsJSON() bool {
	return util.IsMediaTypeJson(r.ContentType)
}

// IsSupported returns true if we support this content type for server. Otherwise io.Reader will be generated
func (r RequestBodyDefinition) IsSupported() bool {
	return r.NameTag != ""
}

// IsFixedContentType returns true if content type has fixed content type, i.e. contains no "*" symbol
func (r RequestBodyDefinition) IsFixedContentType() bool {
	return !strings.Contains(r.ContentType, "*")
}

type RequestBodyEncoding struct {
	ContentType string
	Style       string
	Explode     *bool
}

type ResponseDefinition struct {
	StatusCode  string
	Description string
	Contents    []ResponseContentDefinition
	Headers     []ResponseHeaderDefinition
	Ref         string
}

func (r ResponseDefinition) HasFixedStatusCode() bool {
	_, err := strconv.Atoi(r.StatusCode)
	return err == nil
}

func (r ResponseDefinition) GoName() string {
	return SchemaNameToTypeName(r.StatusCode)
}

func (r ResponseDefinition) IsRef() bool {
	return r.Ref != ""
}

func (r ResponseDefinition) IsExternalRef() bool {
	if !r.IsRef() {
		return false
	}
	return strings.Contains(r.Ref, ".")
}

type ResponseContentDefinition struct {
	// This is the schema describing this content
	Schema Schema

	// This is the content type corresponding to the body, eg, application/json
	ContentType string

	// When we generate type names, we need a Tag for it, such as JSON, in
	// which case we will produce "Response200JSONContent".
	NameTag string
}

// TypeDef returns the Go type definition for a request body
func (r ResponseContentDefinition) TypeDef(opID string, statusCode int) *TypeDefinition {
	return &TypeDefinition{
		TypeName: fmt.Sprintf("%s%v%sResponse", opID, statusCode, r.NameTagOrContentType()),
		Schema:   r.Schema,
	}
}

func (r ResponseContentDefinition) IsSupported() bool {
	return r.NameTag != ""
}

// HasFixedContentType returns true if content type has fixed content type, i.e. contains no "*" symbol
func (r ResponseContentDefinition) HasFixedContentType() bool {
	return !strings.Contains(r.ContentType, "*")
}

func (r ResponseContentDefinition) NameTagOrContentType() string {
	if r.NameTag != "" {
		return r.NameTag
	}
	return SchemaNameToTypeName(r.ContentType)
}

// IsJSON returns whether this is a JSON media type, for instance:
// - application/json
// - application/vnd.api+json
// - application/*+json
func (r ResponseContentDefinition) IsJSON() bool {
	return util.IsMediaTypeJson(r.ContentType)
}

// IsStreamingContentType reports whether this response's media type matches
// any configured streaming-content-types pattern (defaults merged with
// OutputOptions.StreamingContentTypes). Templates use this to emit a
// flush-per-chunk streaming path instead of a buffered io.Copy.
func (r ResponseContentDefinition) IsStreamingContentType() bool {
	for _, re := range globalState.streamingContentTypeRegexes {
		if re.MatchString(r.ContentType) {
			return true
		}
	}
	return false
}

type ResponseHeaderDefinition struct {
	Name              string
	GoName            string
	Schema            Schema
	Required          bool
	Nullable          bool
	Deprecated        bool
	DeprecationReason string
}

// DeprecationComment returns a Go-style deprecation comment if the header is deprecated, otherwise returns an empty string.
func (h ResponseHeaderDefinition) DeprecationComment() string {
	if !h.Deprecated {
		return ""
	}
	reason := h.DeprecationReason
	if reason == "" {
		reason = "this header has been marked as deprecated upstream, but no `x-deprecated-reason` was set"
	}
	return DeprecationComment(reason)
}

// GoTypeDef returns the Go type string for this header, applying pointer or
// nullable wrapping based on the Required/Nullable fields and global config.
func (h ResponseHeaderDefinition) GoTypeDef() string {
	typeDef := h.Schema.TypeDecl()
	if globalState.options.OutputOptions.NullableType && h.Nullable {
		return "nullable.Nullable[" + typeDef + "]"
	}
	if !h.Schema.SkipOptionalPointer && (!h.Required || h.Nullable) {
		typeDef = "*" + typeDef
	}
	return typeDef
}

// IsOptional returns true if this header's Go type is indirect (pointer or
// nullable wrapper), meaning the template should guard before calling
// w.Header().Set(). This must stay in sync with GoTypeDef().
func (h ResponseHeaderDefinition) IsOptional() bool {
	if h.IsNullable() {
		return true
	}
	if h.Schema.SkipOptionalPointer {
		return false
	}
	return !h.Required || h.Nullable
}

// IsNullable returns true if the header type uses nullable.Nullable[T]
// rather than a pointer for optionality.
func (h ResponseHeaderDefinition) IsNullable() bool {
	return globalState.options.OutputOptions.NullableType && h.Nullable
}

// SchemaType returns the first OpenAPI type string for this header's schema
// (e.g. "string", "integer"), or empty string if unavailable.
func (h ResponseHeaderDefinition) SchemaType() string {
	if h.Schema.OAPISchema != nil && h.Schema.OAPISchema.Type != nil {
		if s := h.Schema.OAPISchema.Type.Slice(); len(s) > 0 {
			return s[0]
		}
	}
	return ""
}

// SchemaFormat returns the OpenAPI format string for this header's schema
// (e.g. "date-time", "duration"), or empty string if unavailable.
func (h ResponseHeaderDefinition) SchemaFormat() string {
	if h.Schema.OAPISchema != nil {
		return h.Schema.OAPISchema.Format
	}
	return ""
}

// FilterParameterDefinitionByType returns the subset of the specified parameters which are of the
// specified type.
func FilterParameterDefinitionByType(params []ParameterDefinition, in string) []ParameterDefinition {
	var out []ParameterDefinition
	for _, p := range params {
		if p.In == in {
			out = append(out, p)
		}
	}
	return out
}

// OperationDefinitions returns all operations for a swagger definition.
// pathItemSourceLine returns the 1-based source line on which a path item's
// key is declared, when the spec was loaded with origin tracking enabled
// (see LoadSwagger). Returns 0 when the location is unavailable — e.g. a
// spec built in memory or one whose loader did not record origins.
func pathItemSourceLine(pathItem *openapi3.PathItem) int {
	if pathItem == nil || pathItem.Origin == nil || pathItem.Origin.Key == nil {
		return 0
	}
	return pathItem.Origin.Key.Line
}

func OperationDefinitions(swagger *openapi3.T) ([]OperationDefinition, error) {
	var operations []OperationDefinition

	if swagger == nil || swagger.Paths == nil {
		return operations, nil
	}

	// Collect the names of security schemes actually defined under
	// components/securitySchemes. Requirements that reference an undefined
	// scheme are filtered out below so generated code stays compilable.
	definedSecuritySchemes := map[string]struct{}{}
	if swagger.Components != nil {
		for name := range swagger.Components.SecuritySchemes {
			definedSecuritySchemes[name] = struct{}{}
		}
	}

	// Track alias counters for generating unique client method names
	// when multiple paths $ref the same path item.
	aliasCounters := map[string]int{}

	// Resolve path-item-level (shared) parameters once for the whole spec, so
	// their helper types are declared once per path item and colliding names
	// across paths are disambiguated (issue #2090).
	sharedParams, err := resolveSharedParameters(swagger)
	if err != nil {
		return nil, err
	}

	for _, requestPath := range SortedMapKeys(swagger.Paths.Map()) {
		pathItem := swagger.Paths.Value(requestPath)
		// Source line of this path's key, so route registration can follow
		// spec declaration order (issue #1887). Zero when unavailable.
		pathSpecOrder := pathItemSourceLine(pathItem)
		// Parameters defined for all methods on this path, resolved by the
		// pre-pass. Their helper types are emitted once for the path item (on
		// its first operation) rather than once per operation (issue #2090).
		globalParams := sharedParams[pathItem]
		sharedParamTypeDefs := sharedParameterTypeDefs(globalParams)
		sharedTypeDefsEmitted := false

		// Each path can have a number of operations, POST, GET, OPTIONS, etc.
		pathOps := pathItem.Operations()
		for _, opName := range SortedMapKeys(pathOps) {
			// NOTE that this is a reference to the existing copy of the Operation, so any modifications will modify our shared copy of the spec
			op := pathOps[opName]

			if pathItem.Servers != nil {
				op.Servers = &pathItem.Servers
			}
			// take a copy of operationId, so we don't modify the underlying spec
			operationId := op.OperationID
			// Preserve the raw spec value (pre-normalization, pre-prefix, pre-alias-suffix)
			// so templates that need to mirror the OpenAPI spec verbatim — e.g. echo's
			// per-operation middleware map key — can do so without seeing the
			// Go-identifier-friendly transformations applied below.
			specOperationId := op.OperationID
			// We rely on OperationID to generate function names, it's required
			if operationId == "" {
				operationId, err = generateDefaultOperationID(opName, requestPath)
				if err != nil {
					return nil, fmt.Errorf("error generating default OperationID for %s/%s: %s",
						opName, requestPath, err)
				}
			} else {
				operationId = nameNormalizer(operationId)
			}
			operationId = typeNamePrefix(operationId) + operationId

			// Detect path aliases: when a path item has an internal $ref
			// pointing to another path in the same document (e.g.
			// "#/paths/~1test"), it's a duplicate that would produce
			// identical server methods. External $refs (pointing to other
			// files) are not aliases — they're the sole definition of
			// that path, just stored externally.
			isAlias := strings.HasPrefix(pathItem.Ref, "#/paths/")
			var aliasTarget string
			if isAlias {
				aliasTarget = nameNormalizer(operationId)
				n := aliasCounters[operationId]
				aliasCounters[operationId] = n + 1
				operationId = operationId + fmt.Sprintf("Alias%d", n)
			}

			if !globalState.options.Compatibility.PreserveOriginalOperationIdCasingInEmbeddedSpec && !isAlias {
				// update the existing, shared, copy of the spec if we're not wanting to preserve it.
				// Skip for aliases: they share the same *Operation as the canonical path,
				// and writing the suffixed name back would corrupt the original.
				op.OperationID = operationId
			}

			// These are parameters defined for the specific path method that
			// we're iterating over.
			localParams, err := DescribeParameters(op.Parameters, []string{operationId + "Params"})
			if err != nil {
				return nil, fmt.Errorf("error describing global parameters for %s/%s: %s",
					opName, requestPath, err)
			}
			// All the parameters required by a handler are the union of the
			// global parameters and the local parameters.
			allParams, err := CombineOperationParameters(globalParams, localParams)
			if err != nil {
				return nil, err
			}

			ensureExternalRefsInParameterDefinitions(&allParams, pathItem.Ref)

			// Order the path parameters to match the order as specified in
			// the path, not in the swagger spec, and validate that the parameter
			// names match, as downstream code depends on that.
			pathParams := FilterParameterDefinitionByType(allParams, "path")
			pathParams, err = SortParamsByPath(requestPath, pathParams)
			if err != nil {
				return nil, err
			}

			bodyDefinitions, typeDefinitions, err := GenerateBodyDefinitions(operationId, op.RequestBody, pathItem.Ref)
			if err != nil {
				return nil, fmt.Errorf("error generating body definitions: %w", err)
			}

			ensureExternalRefsInRequestBodyDefinitions(&bodyDefinitions, pathItem.Ref)

			responseDefinitions, err := GenerateResponseDefinitions(operationId, op.Responses.Map(), pathItem.Ref)
			if err != nil {
				return nil, fmt.Errorf("error generating response definitions: %w", err)
			}

			ensureExternalRefsInResponseDefinitions(&responseDefinitions, pathItem.Ref)

			opDef := OperationDefinition{
				PathParams:      pathParams,
				HeaderParams:    FilterParameterDefinitionByType(allParams, "header"),
				QueryParams:     FilterParameterDefinitionByType(allParams, "query"),
				CookieParams:    FilterParameterDefinitionByType(allParams, "cookie"),
				OperationId:     nameNormalizer(operationId),
				SpecOperationId: specOperationId,
				// Replace newlines in summary.
				Summary:         op.Summary,
				Method:          opName,
				Path:            requestPath,
				SpecOrder:       pathSpecOrder,
				Spec:            op,
				Bodies:          bodyDefinitions,
				Responses:       responseDefinitions,
				TypeDefinitions: typeDefinitions,
				IsAlias:         isAlias,
				AliasTarget:     aliasTarget,
				PathItemRef:     pathItem.Ref,
			}

			// check for overrides of SecurityDefinitions.
			// See: "Step 2. Applying security:" from the spec:
			// https://swagger.io/docs/specification/authentication/
			if op.Security != nil {
				opDef.SecurityDefinitions = DescribeSecurityDefinition(*op.Security)
			} else {
				// use global securityDefinitions
				// globalSecurityDefinitions contains the top-level securityDefinitions.
				// They are the default securityPermissions which are injected into each
				// path, except for the case where a path explicitly overrides them.
				opDef.SecurityDefinitions = DescribeSecurityDefinition(swagger.Security)

			}
			opDef.SecurityDefinitions = filterOutUndefinedSecuritySchemes(opDef.SecurityDefinitions, definedSecuritySchemes)

			if op.RequestBody != nil {
				opDef.BodyRequired = op.RequestBody.Value.Required
			}

			if op.Deprecated {
				reason := ""
				if extension, ok := op.Extensions[extDeprecationReason]; ok {
					if r, err := extParseDeprecationReason(extension); err == nil {
						reason = r
					}
				}
				for i := range opDef.Bodies {
					opDef.Bodies[i].Deprecated = true
					opDef.Bodies[i].DeprecationReason = reason
				}
			}

			// Generate all the type definitions needed for this operation
			opDef.TypeDefinitions = append(opDef.TypeDefinitions, GenerateTypeDefsForOperation(opDef)...)

			// Declare the shared (path-item-level) parameter helper types once,
			// on the first operation of the path item (issue #2090).
			if !sharedTypeDefsEmitted {
				opDef.TypeDefinitions = append(opDef.TypeDefinitions, sharedParamTypeDefs...)
				sharedTypeDefsEmitted = true
			}

			operations = append(operations, opDef)
		}
	}
	return operations, nil
}

// WebhookOperationDefinitions extracts OpenAPI 3.1+ webhook operations
// from swagger.Webhooks into the same OperationDefinition shape used for
// path operations, so they flow through the same downstream pipeline
// (body / response generation, type definitions, etc.) but are routed
// to webhook-specific templates.
//
// kin-openapi only populates the Webhooks field for OpenAPI 3.1+
// documents, so a missing/empty map naturally short-circuits this for
// 3.0 specs without an explicit version check.
//
// The result mirrors OperationDefinitions in structure, minus the
// path-alias logic (webhooks have no path template) and the path-
// parameter extraction (webhooks have no path params).
func WebhookOperationDefinitions(swagger *openapi3.T) ([]OperationDefinition, error) {
	var operations []OperationDefinition
	if swagger == nil || len(swagger.Webhooks) == 0 {
		return operations, nil
	}

	sharedParams, err := resolveSharedParameters(swagger)
	if err != nil {
		return nil, err
	}

	for _, webhookName := range SortedMapKeys(swagger.Webhooks) {
		pathItem := swagger.Webhooks[webhookName]
		if pathItem == nil {
			continue
		}

		// Path-item-level parameters apply to every method on the webhook
		// (rare for webhooks, but honored defensively). Their helper types are
		// declared once for the path item (issue #2090).
		globalParams := sharedParams[pathItem]
		sharedParamTypeDefs := sharedParameterTypeDefs(globalParams)
		sharedTypeDefsEmitted := false

		pathOps := pathItem.Operations()
		for _, opName := range SortedMapKeys(pathOps) {
			op := pathOps[opName]

			// Prefer an explicit operationId on the webhook operation;
			// otherwise derive from the webhook map key. Either way,
			// run through the configured name normalizer.
			operationId := op.OperationID
			if operationId == "" {
				operationId = webhookName
			}
			operationId = nameNormalizer(operationId)
			operationId = typeNamePrefix(operationId) + operationId

			localParams, err := DescribeParameters(op.Parameters, []string{operationId + "Params"})
			if err != nil {
				return nil, fmt.Errorf("error describing webhook %q operation params: %w", webhookName, err)
			}
			allParams, err := CombineOperationParameters(globalParams, localParams)
			if err != nil {
				return nil, err
			}

			bodyDefinitions, typeDefinitions, err := GenerateBodyDefinitions(operationId, op.RequestBody, pathItem.Ref)
			if err != nil {
				return nil, fmt.Errorf("error generating body definitions for webhook %q: %w", webhookName, err)
			}

			responseDefinitions, err := GenerateResponseDefinitions(operationId, op.Responses.Map(), pathItem.Ref)
			if err != nil {
				return nil, fmt.Errorf("error generating response definitions for webhook %q: %w", webhookName, err)
			}

			opDef := OperationDefinition{
				HeaderParams:    FilterParameterDefinitionByType(allParams, "header"),
				QueryParams:     FilterParameterDefinitionByType(allParams, "query"),
				CookieParams:    FilterParameterDefinitionByType(allParams, "cookie"),
				OperationId:     nameNormalizer(operationId),
				Summary:         op.Summary,
				Method:          opName,
				Path:            "",
				Spec:            op,
				Bodies:          bodyDefinitions,
				Responses:       responseDefinitions,
				TypeDefinitions: typeDefinitions,
				IsWebhook:       true,
				WebhookName:     webhookName,
			}

			if op.Security != nil {
				opDef.SecurityDefinitions = DescribeSecurityDefinition(*op.Security)
			} else {
				opDef.SecurityDefinitions = DescribeSecurityDefinition(swagger.Security)
			}

			if op.RequestBody != nil {
				opDef.BodyRequired = op.RequestBody.Value.Required
			}

			opDef.TypeDefinitions = append(opDef.TypeDefinitions, GenerateTypeDefsForOperation(opDef)...)
			if !sharedTypeDefsEmitted {
				opDef.TypeDefinitions = append(opDef.TypeDefinitions, sharedParamTypeDefs...)
				sharedTypeDefsEmitted = true
			}
			operations = append(operations, opDef)
		}
	}
	return operations, nil
}

// CallbackOperationDefinitions extracts OpenAPI callback operations
// from spec.Paths.<path>.<method>.Callbacks into the same
// OperationDefinition shape used for path operations, so they flow
// through the same downstream pipeline (body / response generation,
// type definitions, etc.) but are routed to callback-specific
// templates.
//
// Callbacks have been part of the OpenAPI spec since 3.0, so this is
// not gated on version: any spec that declares callbacks gets them
// generated. The spec shape:
//
//	paths:
//	  /api/plant_tree:
//	    post:
//	      operationId: PlantTree
//	      callbacks:
//	        treePlanted:                 # the callback map key
//	          '{$request.body#/url}':    # the runtime URL expression
//	            post:
//	              operationId: TreePlanted
//
// Each leaf operation (the inner `post:` above) becomes one
// OperationDefinition with IsCallback=true and CallbackName set to the
// outer map key ("treePlanted"). The codegen does not interpret the URL
// expression itself -- the caller of the generated CallbackInitiator
// supplies the resolved target URL at runtime (the same way it would
// for a webhook).
func CallbackOperationDefinitions(swagger *openapi3.T) ([]OperationDefinition, error) {
	var operations []OperationDefinition
	if swagger == nil || swagger.Paths == nil {
		return operations, nil
	}

	sharedParams, err := resolveSharedParameters(swagger)
	if err != nil {
		return nil, err
	}

	for _, requestPath := range SortedMapKeys(swagger.Paths.Map()) {
		pathItem := swagger.Paths.Value(requestPath)
		if pathItem == nil {
			continue
		}
		// Iterate path-item operations in sorted method order for
		// deterministic output across runs (Operations() returns a
		// map; range order is randomized).
		parentOps := pathItem.Operations()
		for _, parentMethod := range SortedMapKeys(parentOps) {
			parentOp := parentOps[parentMethod]
			if len(parentOp.Callbacks) == 0 {
				continue
			}
			for _, callbackName := range SortedMapKeys(parentOp.Callbacks) {
				cbRef := parentOp.Callbacks[callbackName]
				if cbRef == nil || cbRef.Value == nil {
					continue
				}
				cb := cbRef.Value
				// A Callback maps URL-expression to PathItem; iterate
				// in sorted key order for deterministic output. The
				// internal map is private so use the accessor pair.
				cbKeys := append([]string(nil), cb.Keys()...)
				slices.Sort(cbKeys)
				for _, urlExpr := range cbKeys {
					cbPathItem := cb.Value(urlExpr)
					if cbPathItem == nil {
						continue
					}
					// Path-item-level parameters shared by every method on
					// the callback path item; helper types declared once
					// (issue #2090).
					globalParams := sharedParams[cbPathItem]
					sharedParamTypeDefs := sharedParameterTypeDefs(globalParams)
					sharedTypeDefsEmitted := false

					cbOps := cbPathItem.Operations()
					for _, opName := range SortedMapKeys(cbOps) {
						op := cbOps[opName]

						operationId := op.OperationID
						if operationId == "" {
							operationId = callbackName
						}
						operationId = nameNormalizer(operationId)
						operationId = typeNamePrefix(operationId) + operationId

						localParams, err := DescribeParameters(op.Parameters, []string{operationId + "Params"})
						if err != nil {
							return nil, fmt.Errorf("error describing callback %q operation params: %w", callbackName, err)
						}
						allParams, err := CombineOperationParameters(globalParams, localParams)
						if err != nil {
							return nil, err
						}

						bodyDefinitions, typeDefinitions, err := GenerateBodyDefinitions(operationId, op.RequestBody, cbPathItem.Ref)
						if err != nil {
							return nil, fmt.Errorf("error generating body definitions for callback %q: %w", callbackName, err)
						}

						responseDefinitions, err := GenerateResponseDefinitions(operationId, op.Responses.Map(), cbPathItem.Ref)
						if err != nil {
							return nil, fmt.Errorf("error generating response definitions for callback %q: %w", callbackName, err)
						}

						opDef := OperationDefinition{
							HeaderParams:    FilterParameterDefinitionByType(allParams, "header"),
							QueryParams:     FilterParameterDefinitionByType(allParams, "query"),
							CookieParams:    FilterParameterDefinitionByType(allParams, "cookie"),
							OperationId:     nameNormalizer(operationId),
							Summary:         op.Summary,
							Method:          opName,
							Path:            "",
							Spec:            op,
							Bodies:          bodyDefinitions,
							Responses:       responseDefinitions,
							TypeDefinitions: typeDefinitions,
							IsCallback:      true,
							CallbackName:    callbackName,
						}

						if op.Security != nil {
							opDef.SecurityDefinitions = DescribeSecurityDefinition(*op.Security)
						} else {
							opDef.SecurityDefinitions = DescribeSecurityDefinition(swagger.Security)
						}

						if op.RequestBody != nil {
							opDef.BodyRequired = op.RequestBody.Value.Required
						}

						opDef.TypeDefinitions = append(opDef.TypeDefinitions, GenerateTypeDefsForOperation(opDef)...)
						if !sharedTypeDefsEmitted {
							opDef.TypeDefinitions = append(opDef.TypeDefinitions, sharedParamTypeDefs...)
							sharedTypeDefsEmitted = true
						}
						operations = append(operations, opDef)
					}
				}
			}
		}
	}
	return operations, nil
}

func generateDefaultOperationID(opName string, requestPath string) (string, error) {
	if opName == "" {
		return "", fmt.Errorf("operation name cannot be an empty string")
	}
	if requestPath == "" {
		return "", fmt.Errorf("request path cannot be an empty string")
	}

	operationID := strings.ToLower(opName)
	for part := range strings.SplitSeq(requestPath, "/") {
		if part != "" {
			operationID = operationID + "-" + part
		}
	}

	return nameNormalizer(operationID), nil
}

// GenerateBodyDefinitions turns the Swagger body definitions into a list of our body
// definitions which will be used for code generation.
//
// pathItemRef is the path item's $ref (if any). When non-empty and pointing at
// an external file, the body type that would otherwise be hoisted locally is
// replaced by a reference to the imported package's same-named type — the
// imported package already declares it (with any As/From/Merge methods), so
// redeclaring locally would just produce an awkward duplicate with
// package-qualified union elements.
func GenerateBodyDefinitions(operationID string, bodyOrRef *openapi3.RequestBodyRef, pathItemRef string) ([]RequestBodyDefinition, []TypeDefinition, error) {
	if bodyOrRef == nil {
		return nil, nil, nil
	}
	body := bodyOrRef.Value

	var bodyDefinitions []RequestBodyDefinition
	var typeDefinitions []TypeDefinition

	for _, contentType := range SortedMapKeys(body.Content) {
		content := body.Content[contentType]
		var tag string
		var defaultBody bool

		switch {
		case contentType == "application/json":
			tag = "JSON"
			defaultBody = true
		case util.IsMediaTypeJson(contentType):
			tag = mediaTypeToCamelCase(contentType)
		case strings.HasPrefix(contentType, "multipart/"):
			tag = "Multipart"
		case contentType == "application/x-www-form-urlencoded":
			tag = "Formdata"
		case contentType == "text/plain":
			tag = "Text"
		default:
			bd := RequestBodyDefinition{
				Required:    body.Required,
				ContentType: contentType,
			}
			bodyDefinitions = append(bodyDefinitions, bd)
			continue
		}

		bodyTypeName := operationID + tag + "Body"
		bodySchema, err := GenerateGoSchema(content.Schema, []string{bodyTypeName})
		if err != nil {
			return nil, nil, fmt.Errorf("error generating request body definition: %w", err)
		}

		// If the body is a pre-defined type
		if content.Schema != nil && IsGoTypeReference(content.Schema.Ref) {
			// Convert the reference path to Go type
			refType, err := RefPathToGoType(content.Schema.Ref)
			if err != nil {
				return nil, nil, fmt.Errorf("error turning reference (%s) into a Go type: %w", content.Schema.Ref, err)
			}
			bodySchema.RefType = refType
		}

		// If the request has a body, but it's not a user defined
		// type under #/components, we'll define a type for it, so
		// that we have an easy to use type for marshaling.
		if bodySchema.RefType == "" {
			if externalPkg := externalPackageFor(pathItemRef); externalPkg != "" {
				// The operation's path item came from an external file; the
				// imported package already declares this body type with the
				// matching name. Reference it instead of redeclaring.
				bodySchema.RefType = fmt.Sprintf("%s.%s", externalPkg, bodyTypeName)
			} else {
				if contentType == "application/x-www-form-urlencoded" {
					// Apply the appropriate structure tag if the request
					// schema was defined under the operations' section.
					for i := range bodySchema.Properties {
						bodySchema.Properties[i].NeedsFormTag = true
					}

					// Regenerate the Golang struct adding the new form tag.
					bodySchema.GoType = GenStructFromSchema(bodySchema)
				}

				td := TypeDefinition{
					TypeName: bodyTypeName,
					Schema:   bodySchema,
				}
				typeDefinitions = append(typeDefinitions, td)
				// The body schema now is a reference to a type
				bodySchema.RefType = bodyTypeName
			}
		}

		bd := RequestBodyDefinition{
			Required:    body.Required,
			Schema:      bodySchema,
			NameTag:     tag,
			ContentType: contentType,
			Default:     defaultBody,
		}

		if len(content.Encoding) != 0 {
			bd.Encoding = make(map[string]RequestBodyEncoding, len(content.Encoding))
			for k, v := range content.Encoding {
				encoding := RequestBodyEncoding{ContentType: v.ContentType, Style: v.Style, Explode: v.Explode}
				bd.Encoding[k] = encoding
			}
		}

		bodyDefinitions = append(bodyDefinitions, bd)
	}
	slices.SortFunc(bodyDefinitions, func(a, b RequestBodyDefinition) int {
		return cmp.Compare(a.ContentType, b.ContentType)
	})
	return bodyDefinitions, typeDefinitions, nil
}

func GenerateResponseDefinitions(operationID string, responses map[string]*openapi3.ResponseRef, pathItemRef string) ([]ResponseDefinition, error) {
	externalPkg := externalPackageFor(pathItemRef)

	var responseDefinitions []ResponseDefinition
	// do not let multiple status codes ref to same response, it will break the type switch
	refSet := make(map[string]struct{})

	for _, statusCode := range SortedMapKeys(responses) {
		responseOrRef := responses[statusCode]
		if responseOrRef == nil {
			continue
		}
		response := responseOrRef.Value

		var responseContentDefinitions []ResponseContentDefinition

		for _, contentType := range SortedMapKeys(response.Content) {
			content := response.Content[contentType]
			var tag string
			switch {
			case contentType == "application/json":
				tag = "JSON"
			case util.IsMediaTypeJson(contentType):
				tag = mediaTypeToCamelCase(contentType)
			case contentType == "application/x-www-form-urlencoded":
				tag = "Formdata"
			case strings.HasPrefix(contentType, "multipart/"):
				tag = "Multipart"
			case contentType == "text/plain":
				tag = "Text"
			default:
				rcd := ResponseContentDefinition{
					ContentType: contentType,
				}
				responseContentDefinitions = append(responseContentDefinitions, rcd)
				continue
			}

			responseTypeName := operationID + statusCode + tag + "Response"
			// The strict-server envelope keeps the bare ...Response name
			// (e.g. "GetPing200JSONResponse"); the hoisted body type is
			// suffixed so the envelope can reference it without colliding
			// (the strict envelope is sometimes a struct that wraps the
			// body in a Body field, which would self-reference if the
			// names matched).
			responseBodyTypeName := responseTypeName + "Body"
			contentSchema, err := GenerateGoSchema(content.Schema, []string{responseBodyTypeName})
			if err != nil {
				return nil, fmt.Errorf("error generating request body definition: %w", err)
			}

			// Hoist inline response-root schemas that need method-emitting
			// boilerplate (UnionElements / AdditionalProperties) to a
			// synthetic top-level TypeDefinition. The hoisted typedef flows
			// via op.TypeDefinitions (collected in
			// GenerateTypeDefsForOperation) and gets declared once via
			// typedef.tmpl with full union/additionalProperties methods.
			// The strict-server template references it as the body type
			// from the envelope.
			//
			// When the operation came from an externally-ref'd path item,
			// the imported package generated the same hoisted name, so we
			// reference it instead of redeclaring locally.
			if !IsGoTypeReference(responseOrRef.Ref) && contentSchema.RefType == "" &&
				(len(contentSchema.UnionElements) != 0 || contentSchema.HasAdditionalProperties ||
					(globalState.options.OutputOptions.GenerateTypesForAnonymousSchemas && len(contentSchema.Properties) > 0)) {
				if externalPkg != "" {
					contentSchema.RefType = fmt.Sprintf("%s.%s", externalPkg, responseBodyTypeName)
				} else {
					contentSchema.AdditionalTypes = append(contentSchema.AdditionalTypes, TypeDefinition{
						TypeName: responseBodyTypeName,
						JsonName: responseBodyTypeName,
						Schema:   contentSchema,
					})
					contentSchema.RefType = responseBodyTypeName
				}
			}

			rcd := ResponseContentDefinition{
				ContentType: contentType,
				NameTag:     tag,
				Schema:      contentSchema,
			}

			responseContentDefinitions = append(responseContentDefinitions, rcd)
		}

		var responseHeaderDefinitions []ResponseHeaderDefinition
		for _, headerName := range SortedMapKeys(response.Headers) {
			header := response.Headers[headerName]
			contentSchema, err := GenerateGoSchema(header.Value.Schema, []string{})
			if err != nil {
				return nil, fmt.Errorf("error generating response header definition: %w", err)
			}
			// When the header component itself is an external `$ref` (it lives
			// in another file) and its schema references a named type, that
			// type is generated into the imported package, not this one. The
			// schema `$ref` is written relative to the external file (e.g.
			// "#/components/schemas/ETagSchema"), so GenerateGoSchema resolved
			// it as a bare local name; qualify it with the header's external
			// package so we emit "externalRef0.ETagSchema" instead of an
			// undefined local "ETagSchema" (issue #2060). Guard on the schema
			// being a reference so an external header with an inline primitive
			// schema (e.g. `type: string`) is not turned into
			// "externalRef0.string".
			if header.Value.Schema != nil && header.Value.Schema.Ref != "" {
				ensureExternalRefsInSchema(&contentSchema, header.Ref)
			}
			var nullable bool
			if header.Value.Schema != nil {
				nullable = schemaIsNullable(header.Value.Schema.Value)
			}
			headerDefinition := ResponseHeaderDefinition{
				Name:     headerName,
				GoName:   SchemaNameToTypeName(headerName),
				Schema:   contentSchema,
				Required: header.Value.Required || globalState.options.Compatibility.HeadersImplicitlyRequired,
				Nullable: nullable,
			}
			if header.Value.Deprecated {
				headerDefinition.Deprecated = true
				if extension, ok := header.Value.Extensions[extDeprecationReason]; ok {
					if r, err := extParseDeprecationReason(extension); err == nil {
						headerDefinition.DeprecationReason = r
					}
				}
			}
			responseHeaderDefinitions = append(responseHeaderDefinitions, headerDefinition)
		}

		rd := ResponseDefinition{
			StatusCode: statusCode,
			Contents:   responseContentDefinitions,
			Headers:    responseHeaderDefinitions,
		}
		if response.Description != nil {
			rd.Description = *response.Description
		}
		if IsGoTypeReference(responseOrRef.Ref) {
			var refType string
			if externalPkg != "" && strings.HasPrefix(responseOrRef.Ref, "#") {
				// The operation came from an externally-ref'd path item and this
				// response ref is relative to that external file, so the
				// referenced response actually lives in the imported package
				// (issue #2308). Resolve the name in the context of the external
				// document: derive it straight from the ref's component name
				// rather than via RefPathToGoType, which resolves "#" refs
				// against the *root* spec and would apply the root spec's
				// collision resolution / x-go-name -- a name the imported
				// package never generated. Refs that already target another file
				// (handled below) are qualified by RefPathToGoType itself.
				refParts := strings.Split(responseOrRef.Ref, "/")
				refType = fmt.Sprintf("%s.%s", externalPkg, SchemaNameToTypeName(refParts[len(refParts)-1]))
			} else {
				// Convert the reference path to Go type
				var err error
				refType, err = RefPathToGoType(responseOrRef.Ref)
				if err != nil {
					return nil, fmt.Errorf("error turning reference (%s) into a Go type: %w", responseOrRef.Ref, err)
				}
			}
			// Check if this ref is already used by another response definition. If not use the ref
			// If we let multiple response definitions alias to same response it will break the type switch
			// so only the first response will use the ref, other will generate new structs
			if _, ok := refSet[refType]; !ok {
				rd.Ref = refType
				refSet[refType] = struct{}{}
			}
			// Ensure content schemas get the external ref qualifier so that
			// non-fixed status code paths (e.g. "default") emit the qualified type.
			for i, rcd := range rd.Contents {
				ensureExternalRefsInSchema(&rcd.Schema, responseOrRef.Ref)
				rd.Contents[i] = rcd
			}
		}
		responseDefinitions = append(responseDefinitions, rd)
	}

	return responseDefinitions, nil
}

func GenerateTypeDefsForOperation(op OperationDefinition) []TypeDefinition {
	var typeDefs []TypeDefinition
	// Start with the params object itself
	if len(op.Params()) != 0 {
		typeDefs = append(typeDefs, GenerateParamsTypes(op)...)
	}

	// Now, go through all the additional types we need to declare. Skip
	// parameters shared at the path-item level: their helper types are
	// declared once for the path item (see sharedParameterTypeDefs), not once
	// per operation, which would redeclare them (issue #2090).
	for _, param := range op.AllParams() {
		if param.Shared {
			continue
		}
		typeDefs = append(typeDefs, param.Schema.AdditionalTypes...)
	}

	for _, body := range op.Bodies {
		typeDefs = append(typeDefs, body.Schema.AdditionalTypes...)
	}

	for _, resp := range op.Responses {
		for _, content := range resp.Contents {
			typeDefs = append(typeDefs, content.Schema.AdditionalTypes...)
		}
	}
	return typeDefs
}

// GenerateParamsTypes defines the schema for a parameters definition object
// which encapsulates all the query, header and cookie parameters for an operation.
func GenerateParamsTypes(op OperationDefinition) []TypeDefinition {
	var typeDefs []TypeDefinition

	objectParams := op.QueryParams
	objectParams = append(objectParams, op.HeaderParams...)
	objectParams = append(objectParams, op.CookieParams...)

	typeName := op.OperationId + "Params"

	s := Schema{}
	for _, param := range objectParams {
		pSchema := param.Schema
		param.Style()
		if pSchema.HasAdditionalProperties {
			propRefName := strings.Join([]string{typeName, param.GoName()}, "_")
			pSchema.RefType = propRefName
			typeDefs = append(typeDefs, TypeDefinition{
				TypeName: propRefName,
				Schema:   param.Schema,
			})
		}
		// Merge extensions, in order of increasing precedence:
		//   1. extensions on the referenced schema (param.Spec.Schema.Value)
		//   2. extensions placed as siblings of a $ref inside the
		//      parameter's schema (param.Spec.Schema.Extensions)
		//   3. extensions on the Parameter object itself
		extensions := make(map[string]any)
		if param.Spec.Schema != nil {
			maps.Copy(extensions, combinedSchemaExtensions(param.Spec.Schema))
		}
		maps.Copy(extensions, param.Spec.Extensions)
		prop := Property{
			Description:   param.Spec.Description,
			JsonFieldName: param.ParamName,
			Required:      param.Required,
			Schema:        pSchema,
			NeedsFormTag:  param.Style() == "form",
			Extensions:    extensions,
			Deprecated:    param.Spec.Deprecated,
		}
		s.Properties = append(s.Properties, prop)
	}

	s.GoType = GenStructFromSchema(s)

	td := TypeDefinition{
		TypeName: typeName,
		Schema:   s,
	}
	return append(typeDefs, td)
}

// GenerateTypesForOperations generates code for all types produced within operations
func GenerateTypesForOperations(t *template.Template, ops []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	addTypes, err := GenerateTemplates([]string{"param-types.tmpl", "request-bodies.tmpl"}, t, ops)
	if err != nil {
		return "", fmt.Errorf("error generating type boilerplate for operations: %w", err)
	}
	if _, err := w.WriteString(addTypes); err != nil {
		return "", fmt.Errorf("error writing boilerplate to buffer: %w", err)
	}

	if err = w.Flush(); err != nil {
		return "", fmt.Errorf("error flushing output buffer for server interface: %w", err)
	}

	return buf.String(), nil
}

// sortOperationsBySpecOrder returns a copy of ops reordered to follow the
// order paths are declared in the spec (by source line, OperationDefinition
// .SpecOrder), with a stable fallback to the incoming order when source
// locations are unavailable (e.g. an in-memory spec) or shared by several
// operations (the methods of one path). This is used only for route
// registration (issue #1887); the rest of the pipeline keeps the sorted
// order so name-collision resolution and type emission are unaffected.
func sortOperationsBySpecOrder(ops []OperationDefinition) []OperationDefinition {
	out := make([]OperationDefinition, len(ops))
	copy(out, ops)
	slices.SortStableFunc(out, func(a, b OperationDefinition) int {
		return a.SpecOrder - b.SpecOrder
	})
	return out
}

// sortOperationsLexicographically returns a copy of ops sorted by path then
// method, reproducing the historical (pre-#1887) route-registration order in
// which OperationDefinitions gathers paths. It is a stable no-op on an
// already-gathered slice, but re-sorts explicitly so the ordering does not
// depend on the caller's input.
func sortOperationsLexicographically(ops []OperationDefinition) []OperationDefinition {
	out := make([]OperationDefinition, len(ops))
	copy(out, ops)
	slices.SortStableFunc(out, func(a, b OperationDefinition) int {
		if c := cmp.Compare(a.Path, b.Path); c != 0 {
			return c
		}
		return cmp.Compare(a.Method, b.Method)
	})
	return out
}

// operationsInRegistrationOrder returns ops in the order route handlers should
// be registered. By default this follows spec-declaration order (issue #1887);
// when the sort-handler-registrations compatibility flag is set it restores the
// historical lexicographic order.
func operationsInRegistrationOrder(ops []OperationDefinition) []OperationDefinition {
	if globalState.options.Compatibility.SortHandlerRegistrations {
		return sortOperationsLexicographically(ops)
	}
	return sortOperationsBySpecOrder(ops)
}

// GenerateIrisServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateIrisServer(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"iris/iris-interface.tmpl", "iris/iris-middleware.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"iris/iris-handler.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateChiServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateChiServer(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"chi/chi-interface.tmpl", "chi/chi-middleware.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"chi/chi-handler.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateFiberServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateFiberServer(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"fiber/fiber-interface.tmpl", "fiber/fiber-middleware.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"fiber/fiber-handler.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateFiberV3Server generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateFiberV3Server(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"fiber-v3/fiber-interface.tmpl", "fiber-v3/fiber-middleware.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"fiber-v3/fiber-handler.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateEchoServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateEchoServer(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"echo/echo-interface.tmpl", "echo/echo-wrappers.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"echo/echo-register.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateEcho5Server generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateEcho5Server(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"echo/v5/echo-interface.tmpl", "echo/v5/echo-wrappers.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"echo/v5/echo-register.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateGinServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateGinServer(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"gin/gin-interface.tmpl", "gin/gin-wrappers.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"gin/gin-register.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateGorillaServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateGorillaServer(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"gorilla/gorilla-interface.tmpl", "gorilla/gorilla-middleware.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"gorilla/gorilla-register.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateStdHTTPServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateStdHTTPServer(t *template.Template, operations []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"stdhttp/std-http-interface.tmpl", "stdhttp/std-http-middleware.tmpl"}, t, operations); err != nil {
		return "", err
	}
	// Route registration follows spec-declaration order (issue #1887).
	if err := GenerateTemplatesIntoBuffer(&buf, []string{"stdhttp/std-http-handler.tmpl"}, t, operationsInRegistrationOrder(operations)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GenerateStrictServer(t *template.Template, operations []OperationDefinition, opts Configuration) (string, error) {

	var templates []string

	if opts.Generate.ChiServer || opts.Generate.GorillaServer || opts.Generate.StdHTTPServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-http.tmpl")
	}
	if opts.Generate.EchoServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-echo.tmpl")
	}
	if opts.Generate.GinServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-gin.tmpl")
	}
	if opts.Generate.FiberServer {
		templates = append(templates, "strict/strict-fiber-interface.tmpl", "strict/strict-fiber.tmpl")
	}
	if opts.Generate.FiberV3Server {
		templates = append(templates, "strict/strict-fiber-v3-interface.tmpl", "strict/strict-fiber-v3.tmpl")
	}
	if opts.Generate.IrisServer {
		templates = append(templates, "strict/strict-iris-interface.tmpl", "strict/strict-iris.tmpl")
	}
	if opts.Generate.Echo5Server {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-echo5.tmpl")
	}

	return GenerateTemplates(templates, t, operations)
}

func GenerateStrictResponses(t *template.Template, responses []ResponseDefinition) (string, error) {
	return GenerateTemplates([]string{"strict/strict-responses.tmpl"}, t, responses)
}

// GenerateClient uses the template engine to generate the function which registers our wrappers
// as Echo path handlers.
func GenerateClient(t *template.Template, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"client.tmpl"}, t, ops)
}

// GenerateClientWithResponses generates a client which extends the basic client which does response
// unmarshaling.
func GenerateClientWithResponses(t *template.Template, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"client-with-responses.tmpl"}, t, ops)
}

// GenerateWebhookInitiator generates the WebhookInitiator -- the
// client-side analog for OpenAPI 3.1 webhooks. It mirrors the path
// Client (struct + options + per-method calls + request builders) but
// takes the target URL per-call instead of from a stored Server field.
// The caller passes only the webhook OperationDefinitions (gathered via
// WebhookOperationDefinitions); path operations are emitted by the
// regular Client templates.
func GenerateWebhookInitiator(t *template.Template, webhookOps []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"webhook-initiator.tmpl"}, t, webhookOps)
}

// GenerateCallbackInitiator generates the CallbackInitiator -- the
// client-side analog for OpenAPI callbacks. Structurally identical to
// GenerateWebhookInitiator but takes the callback OperationDefinitions
// gathered via CallbackOperationDefinitions, which walk paths/operations/
// callbacks rather than spec.Webhooks.
func GenerateCallbackInitiator(t *template.Template, callbackOps []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"callback-initiator.tmpl"}, t, callbackOps)
}

// GenerateStdHTTPReceiver renders the merged stdhttp receiver template
// (used for both webhook and callback receivers). The caller selects
// between them by passing prefix "Webhook" or "Callback" along with the
// matching OperationDefinitions. The template emits a {Prefix}Receiver
// interface plus per-operation {Op}{Prefix}Handler factories with
// query/header parameter binding inline (matching the param-binding
// machinery used by the path-server middleware).
func GenerateStdHTTPReceiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"stdhttp/std-http-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateChiReceiver renders the chi receiver template. Chi shares
// stdhttp's (w, r) handler signature, so the template is structurally
// identical -- only the file path and (in the future, if needed)
// framework-specific helpers differ.
func GenerateChiReceiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"chi/chi-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateGorillaReceiver renders the gorilla/mux receiver template.
// Gorilla shares stdhttp's (w, r) handler signature, so the template
// is structurally identical to stdhttp's.
func GenerateGorillaReceiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"gorilla/gorilla-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateEchoReceiver renders the echo (v4) receiver template. Echo's
// handler shape is `(ctx echo.Context) error`, and binding errors are
// returned via echo.NewHTTPError so echo's framework error chain
// reports them as 400 -- there's no errHandler argument like the
// stdhttp receiver factory has.
func GenerateEchoReceiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"echo/echo-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateEcho5Receiver renders the echo (v5) receiver template. Same
// shape as v4 but with `*echo.Context` (pointer) -- the only API
// difference between echo v4 and v5 that affects the receiver.
func GenerateEcho5Receiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"echo/v5/echo-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateGinReceiver renders the gin receiver template. Gin's handler
// shape is `(c *gin.Context)` (no error return); binding errors abort
// with c.JSON(400, gin.H{"error": ...}). Per-handler middleware is not
// generated here -- gin's idiom prefers engine .Use() composition.
func GenerateGinReceiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"gin/gin-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateFiberReceiver renders the fiber (v2) receiver template.
// Fiber's handler shape is `(c *fiber.Ctx) error`; binding errors are
// returned via fiber.NewError so fiber's error chain reports them as
// 400. Per-handler middleware is not generated; use fiber.App.Use().
func GenerateFiberReceiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"fiber/fiber-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateFiberV3Receiver renders the fiber (v3) receiver template.
// Same shape as v2 but with `fiber.Ctx` (interface, by value) -- the
// only API difference between fiber v2 and v3 that affects the
// receiver.
func GenerateFiberV3Receiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"fiber-v3/fiber-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateIrisReceiver renders the iris receiver template. Iris's
// handler shape is `(ctx iris.Context)` (no error return); binding
// errors set ctx.StatusCode(400) plus ctx.WriteString and return.
// Per-handler middleware is not generated; use app.Use() at the
// engine or Party level.
func GenerateIrisReceiver(t *template.Template, prefix string, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"iris/iris-receiver.tmpl"}, t, NewReceiverTemplateData(prefix, ops))
}

// GenerateTemplatesIntoBuffer executes the named templates against ops and
// writes the results to buf, separating consecutive templates with a newline.
// Rendering into a caller-owned buffer lets a caller compose several passes —
// e.g. feeding spec-ordered operations to the registration template while the
// rest of a server uses the pipeline's default order (issue #1887).
func GenerateTemplatesIntoBuffer(buf *bytes.Buffer, templates []string, t *template.Template, ops any) error {
	for i, tmpl := range templates {
		if i > 0 {
			buf.WriteString("\n")
		}
		if err := t.ExecuteTemplate(buf, tmpl, ops); err != nil {
			return fmt.Errorf("error generating %s: %s", tmpl, err)
		}
	}
	return nil
}

// GenerateTemplates used to generate templates
func GenerateTemplates(templates []string, t *template.Template, ops any) (string, error) {
	var buf bytes.Buffer
	if err := GenerateTemplatesIntoBuffer(&buf, templates, t, ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// CombineOperationParameters combines the Parameters defined at a global level (Parameters defined for all methods on a given path) with the Parameters defined at a local level (Parameters defined for a specific path), preferring the locally defined parameter over the global one
func CombineOperationParameters(globalParams []ParameterDefinition, localParams []ParameterDefinition) ([]ParameterDefinition, error) {
	allParams := make([]ParameterDefinition, 0, len(globalParams)+len(localParams))
	dupCheck := make(map[string]map[string]string)
	for _, p := range localParams {
		if dupCheck[p.In] == nil {
			dupCheck[p.In] = make(map[string]string)
		}
		if _, exist := dupCheck[p.In][p.ParamName]; !exist {
			dupCheck[p.In][p.ParamName] = "local"
			allParams = append(allParams, p)
		} else {
			return nil, fmt.Errorf("duplicate local parameter %s/%s", p.In, p.ParamName)
		}
	}
	for _, p := range globalParams {
		if dupCheck[p.In] == nil {
			dupCheck[p.In] = make(map[string]string)
		}
		if t, exist := dupCheck[p.In][p.ParamName]; !exist {
			dupCheck[p.In][p.ParamName] = "global"
			allParams = append(allParams, p)
		} else if t == "global" {
			return nil, fmt.Errorf("duplicate global parameter %s/%s", p.In, p.ParamName)
		}
	}

	return allParams, nil
}
