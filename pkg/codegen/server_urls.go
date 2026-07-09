package codegen

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

const serverURLPrefix = "ServerUrl"
const serverURLSuffixIterations = 10

// serverURLPlaceholderRE captures `{name}` placeholders in a Server URL
// template. Per OpenAPI 3.0.3 §4.7.7, server-variable names are bound to
// the {name} tokens in `Server Object` URL templates; we treat anything
// matching `{[^/{}]+}` as a placeholder candidate. The character class
// excludes `/` so we don't accidentally span path segments, and `{`/`}`
// so nested braces (which the spec doesn't define anyway) don't merge.
var serverURLPlaceholderRE = regexp.MustCompile(`\{([^/{}]+)\}`)

// urlPlaceholders returns the set of variable names referenced as
// `{name}` placeholders in a Server URL template, deduplicated.
func urlPlaceholders(url string) map[string]struct{} {
	matches := serverURLPlaceholderRE.FindAllStringSubmatch(url, -1)
	if len(matches) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		set[m[1]] = struct{}{}
	}
	return set
}

// ServerObjectDefinition defines the definition of an OpenAPI Server object (https://spec.openapis.org/oas/v3.0.3#server-object) as it is provided to code generation in `oapi-codegen`
type ServerObjectDefinition struct {
	// GoName is the name of the variable for this Server URL
	GoName string

	// OAPISchema is the underlying OpenAPI representation of the Server
	OAPISchema *openapi3.Server
}

// UsedVariables returns the subset of OAPISchema.Variables whose
// `{name}` placeholder actually appears in OAPISchema.URL. Variables
// declared but unused are skipped — they would otherwise produce a
// type, constant, function parameter, and a no-op `strings.ReplaceAll`
// (https://github.com/oapi-codegen/oapi-codegen/issues/2004). Used by
// both server-urls.tmpl and BuildServerURLTypeDefinitions so that
// emitted types and the generated function signature stay in sync.
func (s ServerObjectDefinition) UsedVariables() map[string]*openapi3.ServerVariable {
	if s.OAPISchema == nil || len(s.OAPISchema.Variables) == 0 {
		return nil
	}
	placeholders := urlPlaceholders(s.OAPISchema.URL)
	used := make(map[string]*openapi3.ServerVariable, len(s.OAPISchema.Variables))
	for name, v := range s.OAPISchema.Variables {
		if _, ok := placeholders[name]; ok {
			used[name] = v
		}
	}
	return used
}

// UndeclaredPlaceholders returns the sorted list of `{name}`
// placeholder names that appear in OAPISchema.URL but have no
// corresponding entry in OAPISchema.Variables. The previous code
// generated a function that referenced only declared variables, so
// any undeclared placeholder remained in the URL after substitution
// and the trailing `{`/`}` runtime check tripped on every call —
// making the generated function permanently unusable
// (https://github.com/oapi-codegen/oapi-codegen/issues/2005). The
// template now adds these as plain `string` parameters so callers
// can fill them in directly.
func (s ServerObjectDefinition) UndeclaredPlaceholders() []string {
	if s.OAPISchema == nil {
		return nil
	}
	placeholders := urlPlaceholders(s.OAPISchema.URL)
	if len(placeholders) == 0 {
		return nil
	}
	var undeclared []string
	for name := range placeholders {
		if _, declared := s.OAPISchema.Variables[name]; !declared {
			undeclared = append(undeclared, name)
		}
	}
	if len(undeclared) == 0 {
		return nil
	}
	sort.Strings(undeclared)
	return undeclared
}

// serverObjectDefinitions deconflicts server names and returns the
// stable, deterministically-named ServerObjectDefinitions for `spec`.
// Used by both BuildServerURLTypeDefinitions and GenerateServerURLs so
// they generate identifiers that match.
func serverObjectDefinitions(spec *openapi3.T) ([]ServerObjectDefinition, error) {
	names := make(map[string]*openapi3.Server)

	for _, server := range spec.Servers {
		var name string
		if goNameExt, ok := server.Extensions[extGoName]; ok {
			customName, err := extParseGoFieldName(goNameExt)
			if err != nil {
				return nil, fmt.Errorf("invalid value for %q: %w", extGoName, err)
			}
			if customName != "" {
				name = customName
			}
		}
		if name == "" {
			suffix := server.Description
			if suffix == "" {
				suffix = nameNormalizer(server.URL)
			}
			name = serverURLPrefix + UppercaseFirstCharacter(suffix)
			name = nameNormalizer(name)
		}

		// if this is the only type with this name, store it
		if _, conflict := names[name]; !conflict {
			names[name] = server
			continue
		}

		// otherwise, try appending a number to the name. Start at 1 so
		// `Foo` / `Foo1` reads better than `Foo` / `Foo0`.
		saved := false
		for i := 1; i < 1+serverURLSuffixIterations; i++ {
			suffixed := name + strconv.Itoa(i)
			if _, suffixConflict := names[suffixed]; !suffixConflict {
				names[suffixed] = server
				saved = true
				break
			}
		}

		if saved {
			continue
		}

		return nil, fmt.Errorf("failed to create a unique name for the Server URL (%#v) with description (%#v) after %d iterations", server.URL, server.Description, serverURLSuffixIterations)
	}

	keys := SortedMapKeys(names)
	servers := make([]ServerObjectDefinition, len(keys))
	for i, k := range keys {
		servers[i] = ServerObjectDefinition{
			GoName:     k,
			OAPISchema: names[k],
		}
	}
	return servers, nil
}

// serverURLVariableTypeName returns the Go type identifier for the
// `enum`-typed variable `varName` on the server with deconflicted Go
// name `serverGoName`. Mirrors the naming scheme used by
// server-urls.tmpl (`<server>%sVariable`) so that synthesized
// TypeDefinitions, the function signature emitted by the template,
// and any user references all resolve to the same identifier.
func serverURLVariableTypeName(serverGoName, varName string) string {
	return serverGoName + UppercaseFirstCharacter(varName) + "Variable"
}

// serverURLEnumKeys returns deterministic identifier suffixes for each
// enum value of `v`, in `v.Enum` order. Each suffix is just the value
// with its first character upper-cased — matching what the previous
// template-only path produced for happy-path specs (`<Prefix><Value | ucFirst>`),
// so adopters with non-colliding enums keep their existing identifiers.
//
// When two values fold to the same suffix (e.g. `enum: ["foo", "FOO"]`,
// both → `Foo`), later occurrences get a numeric suffix (`Foo`, `Foo1`,
// `Foo2`, …) — same scheme as `SanitizeEnumNames`. The previous
// template path didn't dedup at all and would have produced a
// duplicate-const compile error for these specs; this is purely a
// correctness improvement, not a naming change for any spec that
// actually compiled before.
func serverURLEnumKeys(v *openapi3.ServerVariable) []string {
	keys := make([]string, len(v.Enum))
	seen := make(map[string]int, len(v.Enum))
	for i, val := range v.Enum {
		base := UppercaseFirstCharacter(val)
		if n, dup := seen[base]; dup {
			keys[i] = base + strconv.Itoa(n)
			seen[base] = n + 1
		} else {
			keys[i] = base
			seen[base] = 1
		}
	}
	return keys
}

// serverURLEnumValues returns the EnumValues map handed to GenerateEnums
// for variable `v`: key is the deduplicated identifier suffix from
// serverURLEnumKeys, value is the original enum string.
func serverURLEnumValues(v *openapi3.ServerVariable) map[string]string {
	keys := serverURLEnumKeys(v)
	out := make(map[string]string, len(keys))
	for i, k := range keys {
		out[k] = v.Enum[i]
	}
	return out
}

// ServerURLDefaultPointer carries the pre-computed identifier names
// needed to emit a typed default-pointer constant for an enum-typed
// server-URL variable. Computed in Go (not the template) so that the
// pointer's reference to the enum-value constant always agrees with
// the identifier the const-block actually emitted, including in
// dedup-suffix cases (e.g. `enum: ["foo", "FOO"]` with `default: "FOO"`
// → enum const is `<Prefix>Foo1`, pointer must reference exactly that).
type ServerURLDefaultPointer struct {
	// VariableName is the OpenAPI variable name (for doc comments).
	VariableName string
	// TypeName is the enum's Go type, e.g. `ServerUrlFooBarVariable`.
	TypeName string
	// PointerName is the constant being declared. Normally it is
	// `<TypeName>Default` (matching the historical naming). Switches
	// to `<TypeName>DefaultValue` only when the variable's enum
	// contains a value whose identifier suffix is the literal string
	// "Default" — i.e. exactly the case that produced #2003's
	// duplicate-const compile error under the old codegen. Specs that
	// compiled cleanly before keep their `Default` name.
	PointerName string
	// TargetName is the fully-qualified enum-value constant that the
	// pointer references, e.g. `<TypeName>N443` or `<TypeName>Foo1`.
	TargetName string
	// Description is the variable's OpenAPI description (doc comment).
	Description string
}

// NewServerFunctionParams returns the formatted parameter list for
// the generated `New<ServerName>(...)` function — one typed parameter
// per declared-and-used variable, plus one `string` parameter per
// `{name}` placeholder that appears in the URL but isn't in
// `variables` (#2005). The two groups are sorted together
// alphabetically so the generated signature is deterministic.
//
// Equivalent to calling `genServerURLWithVariablesFunctionParams` for
// the typed half and concatenating the undeclared half — but exposing
// it as a method keeps server-urls.tmpl free of that combination
// logic and means the template helper itself stayed at its
// pre-existing two-argument shape (so any user-supplied custom
// `server-urls.tmpl` override that calls the helper directly is
// unaffected).
func (s ServerObjectDefinition) NewServerFunctionParams() string {
	used := s.UsedVariables()
	undeclared := s.UndeclaredPlaceholders()
	if len(used) == 0 && len(undeclared) == 0 {
		return ""
	}

	type param struct {
		name string
		typ  string
	}
	parts := make([]param, 0, len(used)+len(undeclared))
	for _, k := range SortedMapKeys(used) {
		parts = append(parts, param{
			name: k,
			typ:  serverURLVariableTypeName(s.GoName, k),
		})
	}
	for _, k := range undeclared {
		parts = append(parts, param{name: k, typ: "string"})
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].name < parts[j].name })

	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = p.name + " " + p.typ
	}
	return strings.Join(out, ", ")
}

// EnumDefaultPointers returns one entry per enum-typed used variable
// with a `default` set. The template iterates this directly rather
// than recomputing identifier names from `$v.Default`.
func (s ServerObjectDefinition) EnumDefaultPointers() []ServerURLDefaultPointer {
	used := s.UsedVariables()
	var out []ServerURLDefaultPointer
	for _, varName := range SortedMapKeys(used) {
		v := used[varName]
		if v == nil || len(v.Enum) == 0 || v.Default == "" {
			continue
		}
		keys := serverURLEnumKeys(v)
		var targetKey string
		for i, val := range v.Enum {
			if val == v.Default {
				targetKey = keys[i]
				break
			}
		}
		if targetKey == "" {
			// `default` not in `enum`; BuildServerURLTypeDefinitions
			// already errors on this at codegen time, so we'd never
			// actually reach the template emission. Skip defensively.
			continue
		}
		prefix := serverURLVariableTypeName(s.GoName, varName)
		pointerName := prefix + "Default"
		for _, k := range keys {
			if k == "Default" {
				pointerName = prefix + "DefaultValue"
				break
			}
		}
		out = append(out, ServerURLDefaultPointer{
			VariableName: varName,
			TypeName:     prefix,
			PointerName:  pointerName,
			TargetName:   prefix + targetKey,
			Description:  v.Description,
		})
	}
	return out
}

// BuildServerURLTypeDefinitions synthesizes a TypeDefinition for every
// server-URL variable that defines an `enum`. These are appended into
// the same TypeDefinition slices used by GenerateTypes (typedef.tmpl)
// and GenerateEnums (constants.tmpl), so that server-URL enum
// variables get the same `type X string`, `const ( … )` block, and
// `Valid()` method as any other enum-bearing schema. The
// server-urls.tmpl template no longer emits these declarations
// directly.
//
// Variables without an `enum` are not handled here: server-urls.tmpl
// continues to emit their `type` and (optional) default constant
// inline, since the generic enum path has nothing to contribute for a
// non-enum string.
func BuildServerURLTypeDefinitions(spec *openapi3.T) ([]TypeDefinition, error) {
	servers, err := serverObjectDefinitions(spec)
	if err != nil {
		return nil, err
	}

	var defs []TypeDefinition
	for _, srv := range servers {
		used := srv.UsedVariables()
		// Iterate variables in deterministic order.
		for _, varName := range SortedMapKeys(used) {
			v := used[varName]
			if v == nil || len(v.Enum) == 0 {
				continue
			}
			// Validate that `default`, if set, is one of the
			// declared `enum` values. Per OpenAPI 3.0.3 §4.7.10,
			// the default MUST be in the enum list when both are
			// present; the previous template trusted the spec
			// blindly and emitted a const referencing an
			// undeclared identifier when the spec violated the
			// rule, producing a confusing user-side compile error
			// (https://github.com/oapi-codegen/oapi-codegen/issues/2007).
			// Catch the violation at codegen time so the user sees
			// a clear message pointing at their spec.
			if v.Default != "" {
				inEnum := false
				for _, ev := range v.Enum {
					if ev == v.Default {
						inEnum = true
						break
					}
				}
				if !inEnum {
					return nil, fmt.Errorf("server URL %q: variable %q has default value %q which is not one of the declared enum values %v",
						srv.OAPISchema.URL, varName, v.Default, v.Enum)
				}
			}
			typeName := serverURLVariableTypeName(srv.GoName, varName)
			enumValues := serverURLEnumValues(v)
			defs = append(defs, TypeDefinition{
				TypeName: typeName,
				JsonName: varName,
				Schema: Schema{
					GoType:      "string",
					EnumValues:  enumValues,
					Description: v.Description,
				},
				ForceEnumPrefix: true,
			})
		}
	}
	return defs, nil
}

func GenerateServerURLs(t *template.Template, spec *openapi3.T) (string, error) {
	servers, err := serverObjectDefinitions(spec)
	if err != nil {
		return "", err
	}
	return GenerateTemplates([]string{"server-urls.tmpl"}, t, servers)
}
