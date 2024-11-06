package codegen

import (
	"errors"
	"fmt"
	"reflect"
)

type AdditionalImport struct {
	Alias   string `yaml:"alias,omitempty"`
	Package string `yaml:"package"`
}

// Configuration defines code generation customizations
type Configuration struct {
	// PackageName to generate the code under
	PackageName string `yaml:"package"`
	// Generate specifies which supported output formats to generate
	Generate GenerateOptions `yaml:"generate,omitempty"`
	// CompatibilityOptions specifies backward compatibility settings for the code generator
	Compatibility CompatibilityOptions `yaml:"compatibility,omitempty"`
	// OutputOptions are used to modify the output code in some way.
	OutputOptions OutputOptions `yaml:"output-options,omitempty"`
	// ImportMapping specifies the golang package path for each external reference
	ImportMapping map[string]string `yaml:"import-mapping,omitempty"`
	// AdditionalImports defines any additional Go imports to add to the generated code
	AdditionalImports []AdditionalImport `yaml:"additional-imports,omitempty"`
	// NoVCSVersionOverride allows overriding the version of the application for cases where no Version Control System (VCS) is available when building, for instance when using a Nix derivation.
	// See documentation for how to use it in examples/no-vcs-version-override/README.md
	NoVCSVersionOverride *string `yaml:"-"`
}

// Validate checks whether Configuration represent a valid configuration
func (o Configuration) Validate() error {
	if o.PackageName == "" {
		return errors.New("package name must be specified")
	}

	// Only one server type should be specified at a time.
	nServers := 0
	if o.Generate.IrisServer {
		nServers++
	}
	if o.Generate.ChiServer {
		nServers++
	}
	if o.Generate.FiberServer {
		nServers++
	}
	if o.Generate.EchoServer {
		nServers++
	}
	if o.Generate.GorillaServer {
		nServers++
	}
	if o.Generate.StdHTTPServer {
		nServers++
	}
	if o.Generate.GinServer {
		nServers++
	}
	if nServers > 1 {
		return errors.New("only one server type is supported at a time")
	}

	var errs []error
	if problems := o.Generate.Validate(); problems != nil {
		for k, v := range problems {
			errs = append(errs, fmt.Errorf("`generate` configuration for %v was incorrect: %v", k, v))
		}
	}

	if problems := o.Compatibility.Validate(); problems != nil {
		for k, v := range problems {
			errs = append(errs, fmt.Errorf("`compatibility-options` configuration for %v was incorrect: %v", k, v))
		}
	}

	if problems := o.OutputOptions.Validate(); problems != nil {
		for k, v := range problems {
			errs = append(errs, fmt.Errorf("`output-options` configuration for %v was incorrect: %v", k, v))
		}
	}

	err := errors.Join(errs...)
	if err != nil {
		return fmt.Errorf("failed to validate configuration: %w", err)
	}

	return nil
}

// UpdateDefaults sets reasonable default values for unset fields in Configuration
func (o Configuration) UpdateDefaults() Configuration {
	if reflect.ValueOf(o.Generate).IsZero() {
		o.Generate = GenerateOptions{
			EchoServer:   true,
			Models:       true,
			EmbeddedSpec: true,
		}
	}
	return o
}

// GenerateOptions specifies which supported output formats to generate.
type GenerateOptions struct {
	// IrisServer specifies whether to generate iris server boilerplate
	IrisServer bool `yaml:"iris-server,omitempty"`
	// ChiServer specifies whether to generate chi server boilerplate
	ChiServer bool `yaml:"chi-server,omitempty"`
	// FiberServer specifies whether to generate fiber server boilerplate
	FiberServer bool `yaml:"fiber-server,omitempty"`
	// EchoServer specifies whether to generate echo server boilerplate
	EchoServer bool `yaml:"echo-server,omitempty"`
	// GinServer specifies whether to generate gin server boilerplate
	GinServer bool `yaml:"gin-server,omitempty"`
	// GorillaServer specifies whether to generate Gorilla server boilerplate
	GorillaServer bool `yaml:"gorilla-server,omitempty"`
	// StdHTTPServer specifies whether to generate stdlib http server boilerplate
	StdHTTPServer bool `yaml:"std-http-server,omitempty"`
	// Strict specifies whether to generate strict server wrapper
	Strict bool `yaml:"strict-server,omitempty"`
	// Client specifies whether to generate client boilerplate
	Client bool `yaml:"client,omitempty"`
	// Models specifies whether to generate type definitions
	Models bool `yaml:"models,omitempty"`
	// EmbeddedSpec indicates whether to embed the swagger spec in the generated code
	EmbeddedSpec bool `yaml:"embedded-spec,omitempty"`
}

func (oo GenerateOptions) Validate() map[string]string {
	return nil
}

// CompatibilityOptions specifies backward compatibility settings for the
// code generator.
type CompatibilityOptions struct {
	// In the past, we merged schemas for `allOf` by inlining each schema
	// within the schema list. This approach, though, is incorrect because
	// `allOf` merges at the schema definition level, not at the resulting model
	// level. So, new behavior merges OpenAPI specs but generates different code
	// than we have in the past. Set OldMergeSchemas to true for the old behavior.
	// Please see https://github.com/oapi-codegen/oapi-codegen/issues/531
	OldMergeSchemas bool `yaml:"old-merge-schemas,omitempty"`
	// Enum values can generate conflicting typenames, so we've updated the
	// code for enum generation to avoid these conflicts, but it will result
	// in some enum types being renamed in existing code. Set OldEnumConflicts to true
	// to revert to old behavior. Please see:
	// Please see https://github.com/oapi-codegen/oapi-codegen/issues/549
	OldEnumConflicts bool `yaml:"old-enum-conflicts,omitempty"`
	// It was a mistake to generate a go type definition for every $ref in
	// the OpenAPI schema. New behavior uses type aliases where possible, but
	// this can generate code which breaks existing builds. Set OldAliasing to true
	// for old behavior.
	// Please see https://github.com/oapi-codegen/oapi-codegen/issues/549
	OldAliasing bool `yaml:"old-aliasing,omitempty"`
	// When an object contains no members, and only an additionalProperties specification,
	// it is flattened to a map
	DisableFlattenAdditionalProperties bool `yaml:"disable-flatten-additional-properties,omitempty"`
	// When an object property is both required and readOnly the go model is generated
	// as a pointer. Set DisableRequiredReadOnlyAsPointer to true to mark them as non pointer.
	// Please see https://github.com/oapi-codegen/oapi-codegen/issues/604
	DisableRequiredReadOnlyAsPointer bool `yaml:"disable-required-readonly-as-pointer,omitempty"`
	// When set to true, always prefix enum values with their type name instead of only
	// when typenames would be conflicting.
	AlwaysPrefixEnumValues bool `yaml:"always-prefix-enum-values,omitempty"`
	// Our generated code for Chi has historically inverted the order in which Chi middleware is
	// applied such that the last invoked middleware ends up executing first in the Chi chain
	// This resolves the behavior such that middlewares are chained in the order they are invoked.
	// Please see https://github.com/oapi-codegen/oapi-codegen/issues/786
	ApplyChiMiddlewareFirstToLast bool `yaml:"apply-chi-middleware-first-to-last,omitempty"`
	// Our generated code for gorilla/mux has historically inverted the order in which gorilla/mux middleware is
	// applied such that the last invoked middleware ends up executing first in the middlewares chain
	// This resolves the behavior such that middlewares are chained in the order they are invoked.
	// Please see https://github.com/oapi-codegen/oapi-codegen/issues/841
	ApplyGorillaMiddlewareFirstToLast bool `yaml:"apply-gorilla-middleware-first-to-last,omitempty"`
	// CircularReferenceLimit allows controlling the limit for circular reference checking.
	// In some OpenAPI specifications, we have a higher number of circular
	// references than is allowed out-of-the-box, but can be tuned to allow
	// traversing them.
	// Deprecated: In kin-openapi v0.126.0 (https://github.com/getkin/kin-openapi/tree/v0.126.0?tab=readme-ov-file#v01260) the Circular Reference Counter functionality was removed, instead resolving all references with backtracking, to avoid needing to provide a limit to reference counts.
	CircularReferenceLimit int `yaml:"circular-reference-limit"`
	// AllowUnexportedStructFieldNames makes it possible to output structs that have fields that are unexported.
	//
	// This is expected to be used in conjunction with `x-go-name` and `x-oapi-codegen-only-honour-go-name` to override the resulting output field name, and `x-oapi-codegen-extra-tags` to not produce JSON tags for `encoding/json`, such as:
	//
	//  ```yaml
	//   id:
	//     type: string
	//     x-go-name: accountIdentifier
	//     x-oapi-codegen-extra-tags:
	//       json: "-"
	//     x-oapi-codegen-only-honour-go-name: true
	//   ```
	//
	// NOTE that this can be confusing to users of your OpenAPI specification, who may see a field present and therefore be expecting to see/use it in the request/response, without understanding the nuance of how `oapi-codegen` generates the code.
	AllowUnexportedStructFieldNames bool `yaml:"allow-unexported-struct-field-names"`
}

func (co CompatibilityOptions) Validate() map[string]string {
	return nil
}

// OutputOptions are used to modify the output code in some way.
type OutputOptions struct {
	// Whether to skip go imports on the generated code
	SkipFmt bool `yaml:"skip-fmt,omitempty"`
	// Whether to skip pruning unused components on the generated code
	SkipPrune bool `yaml:"skip-prune,omitempty"`
	// Only include operations that have one of these tags. Ignored when empty.
	IncludeTags []string `yaml:"include-tags,omitempty"`
	// Exclude operations that have one of these tags. Ignored when empty.
	ExcludeTags []string `yaml:"exclude-tags,omitempty"`
	// Only include operations that have one of these operation-ids. Ignored when empty.
	IncludeOperationIDs []string `yaml:"include-operation-ids,omitempty"`
	// Exclude operations that have one of these operation-ids. Ignored when empty.
	ExcludeOperationIDs []string `yaml:"exclude-operation-ids,omitempty"`
	// Override built-in templates from user-provided files
	UserTemplates map[string]string `yaml:"user-templates,omitempty"`

	// Exclude from generation schemas with given names. Ignored when empty.
	ExcludeSchemas []string `yaml:"exclude-schemas,omitempty"`
	// The suffix used for responses types
	ResponseTypeSuffix string `yaml:"response-type-suffix,omitempty"`
	// Override the default generated client type with the value
	ClientTypeName string `yaml:"client-type-name,omitempty"`
	// Whether to use the initialism overrides
	InitialismOverrides bool `yaml:"initialism-overrides,omitempty"`
	// Whether to generate nullable type for nullable fields
	NullableType bool `yaml:"nullable-type,omitempty"`

	// DisableTypeAliasesForType allows defining which OpenAPI `type`s will explicitly not use type aliases
	// Currently supports:
	//   "array"
	DisableTypeAliasesForType []string `yaml:"disable-type-aliases-for-type"`

	// NameNormalizer is the method used to normalize Go names and types, for instance converting the text `MyApi` to `MyAPI`. Corresponds with the constants defined for `codegen.NameNormalizerFunction`
	NameNormalizer string `yaml:"name-normalizer,omitempty"`

	// Overlay defines configuration for the OpenAPI Overlay (https://github.com/OAI/Overlay-Specification) to manipulate the OpenAPI specification before generation. This allows modifying the specification without needing to apply changes directly to it, making it easier to keep it up-to-date.
	Overlay OutputOptionsOverlay `yaml:"overlay"`
}

func (oo OutputOptions) Validate() map[string]string {
	return nil
}

type OutputOptionsOverlay struct {
	Path string `yaml:"path"`

	// Strict defines whether the Overlay should be applied in a strict way, highlighting any actions that will not take any effect. This can, however, lead to more work when testing new actions in an Overlay, so can be turned off with this setting.
	// Defaults to true.
	Strict *bool `yaml:"strict,omitempty"`
}
