package codegen

import (
	"fmt"
	"go/token"
	"strings"
)

// NamingOptions holds opt-in adjustments to how Go names are derived from the
// names in a spec. Each one switches a single, narrowly scoped behavior of the
// configured `name-normalizer`; with none of them set, names are derived
// exactly as they always were.
type NamingOptions struct {
	// EnumValues adjusts the names of the constants generated for enum values.
	EnumValues EnumValueNamingOptions `yaml:"enum-values,omitempty"`
	// Properties adjusts the names of the struct fields generated for object
	// properties and for the parameters of an operation.
	Properties PropertyNamingOptions `yaml:"properties,omitempty"`
	// ExplicitNames decides what happens to a Go name the spec states itself,
	// with `x-go-name` on a property or a parameter, or with `x-enum-varnames`
	// / `x-enumNames`. By default such a name goes through the name normalizer
	// like any other. Set to `verbatim` to emit it exactly as written.
	ExplicitNames ExplicitNamesMode `yaml:"explicit-names,omitempty"`
}

// EnumValueNamingOptions adjusts the names of enum constants.
type EnumValueNamingOptions struct {
	// KeepUnderscores stops the name normalizer from treating an underscore as
	// a word separator that it removes. The words on either side are
	// normalized as usual and the underscore between them is kept, so the enum
	// value `SOME_ENUM_NAME` becomes the constant `SOME_ENUM_NAME` rather than
	// `SOMEENUMNAME`.
	KeepUnderscores bool `yaml:"keep-underscores,omitempty"`
}

// PropertyNamingOptions adjusts the names of struct fields: those of an
// object's properties, and those of an operation's parameters.
type PropertyNamingOptions struct {
	// LeadingUnderscore decides what the underscores a property or parameter
	// name starts with contribute to its Go name. By default each one is spelled out,
	// so `_id` becomes `UnderscoreId`. Set to `drop` for them to contribute
	// nothing, so `_id` becomes `Id`.
	LeadingUnderscore LeadingUnderscoreMode `yaml:"leading-underscore,omitempty"`
}

// ExplicitNamesMode is a value of `output-options.naming.explicit-names`.
type ExplicitNamesMode string

const (
	// ExplicitNamesNormalize is the default: a name the spec states itself
	// goes through the name normalizer.
	ExplicitNamesNormalize ExplicitNamesMode = ""
	// ExplicitNamesVerbatim emits a name the spec states itself exactly as
	// written.
	ExplicitNamesVerbatim ExplicitNamesMode = "verbatim"
)

// LeadingUnderscoreMode is a value of
// `output-options.naming.properties.leading-underscore`.
type LeadingUnderscoreMode string

const (
	// LeadingUnderscoreSpell is the default: a leading underscore becomes the
	// word `Underscore`.
	LeadingUnderscoreSpell LeadingUnderscoreMode = ""
	// LeadingUnderscoreDrop makes leading underscores contribute nothing to
	// the name.
	LeadingUnderscoreDrop LeadingUnderscoreMode = "drop"
)

// Validate reports the naming options that have a value they do not accept,
// keyed by option.
func (n NamingOptions) Validate() map[string]string {
	switch n.ExplicitNames {
	case ExplicitNamesNormalize, ExplicitNamesVerbatim:
	default:
		return map[string]string{
			"naming.explicit-names": fmt.Sprintf("%q is not a known value; the only one is %q", n.ExplicitNames, ExplicitNamesVerbatim),
		}
	}
	switch n.Properties.LeadingUnderscore {
	case LeadingUnderscoreSpell, LeadingUnderscoreDrop:
	default:
		return map[string]string{
			"naming.properties.leading-underscore": fmt.Sprintf("%q is not a known value; the only one is %q", n.Properties.LeadingUnderscore, LeadingUnderscoreDrop),
		}
	}
	return nil
}

// explicitNamesVerbatim reports whether a Go name the spec states itself is
// emitted as written rather than normalized.
func explicitNamesVerbatim() bool {
	return globalState.options.OutputOptions.Naming.ExplicitNames == ExplicitNamesVerbatim
}

// enumValueNameToGoName is SchemaNameToTypeName for the name of an enum
// constant: the same conversion, adjusted by the enum-value naming options.
func enumValueNameToGoName(name string) string {
	if !globalState.options.OutputOptions.Naming.EnumValues.KeepUnderscores {
		return SchemaNameToTypeName(name)
	}
	return typeNamePrefix(name) + normalizeKeepingUnderscores(name)
}

// normalizeKeepingUnderscores applies the name normalizer to each of the
// underscore-separated words of name and joins the results with the
// underscores it would otherwise have removed. Working a word at a time keeps
// this independent of how any one normalizer treats its separators.
func normalizeKeepingUnderscores(name string) string {
	words := strings.Split(name, "_")
	for i, word := range words {
		words[i] = nameNormalizer(word)
	}
	return strings.Join(words, "_")
}

// propertyNameToGoName is SchemaNameToTypeName for the name of a struct field,
// adjusted by the property naming options. A parameter is named with it too:
// it is a field of the operation's parameters struct, and the code that reads
// the field has to arrive at the same name.
func propertyNameToGoName(name string) string {
	if globalState.options.OutputOptions.Naming.Properties.LeadingUnderscore == LeadingUnderscoreDrop {
		// A name that is nothing but underscores keeps them: there is no
		// other name to derive a field from.
		if trimmed := strings.TrimLeft(name, "_"); trimmed != "" {
			name = trimmed
		}
	}
	return SchemaNameToTypeName(name)
}

// checkVerbatimEnumNames validates the names of an enum that are emitted as
// written. Such a name reaches the generated code without being sanitized, so
// it has to be a Go identifier; and since nothing renames it, the same name
// given to two values would silently lose one of them.
func checkVerbatimEnumNames(names []string, verbatim []bool) error {
	seen := make(map[string]struct{}, len(names))
	for i, name := range names {
		if !verbatim[i] {
			continue
		}
		if !token.IsIdentifier(name) {
			return fmt.Errorf("enum name %q from %s or %s must be a valid Go identifier, as output-options.naming.explicit-names is %q", name, extEnumVarNames, extEnumNames, ExplicitNamesVerbatim)
		}
		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf("%s or %s names two values %q, and output-options.naming.explicit-names is %q, so neither is renamed", extEnumVarNames, extEnumNames, name, ExplicitNamesVerbatim)
		}
		seen[name] = struct{}{}
	}
	return nil
}
