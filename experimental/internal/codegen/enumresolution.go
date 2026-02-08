package codegen

import "fmt"

// EnumInfo holds all the information needed to generate an enum's constant block.
// It is populated during the pre-pass phase of code generation.
type EnumInfo struct {
	// TypeName is the Go type name for the enum (e.g., "Color", "Status").
	TypeName string
	// BaseType is the Go base type (e.g., "string", "int").
	BaseType string
	// Values are the raw enum values from the spec.
	Values []string
	// CustomNames are user-provided constant names from x-oapi-codegen-enum-var-names.
	// May be nil or shorter than Values.
	CustomNames []string
	// Doc is the enum's documentation string.
	Doc string
	// SchemaPath is the key used to look up this EnumInfo (schema path string).
	SchemaPath string

	// PrefixTypeName indicates whether constant names should be prefixed with the type name.
	// Set by resolveEnumCollisions when collisions are detected.
	PrefixTypeName bool
	// SanitizedNames are the computed constant name suffixes (without type prefix).
	// Populated by computeEnumConstantNames.
	SanitizedNames []string
}

// finalConstName returns the actual constant name that will appear in generated code.
func (info *EnumInfo) finalConstName(i int) string {
	if info.PrefixTypeName {
		return info.TypeName + info.SanitizedNames[i]
	}
	return info.SanitizedNames[i]
}

// computeEnumConstantNames populates SanitizedNames for each EnumInfo using the
// NameConverter pipeline. Within each enum, duplicate names get numeric suffixes.
func computeEnumConstantNames(infos []*EnumInfo, converter *NameConverter) {
	for _, info := range infos {
		names := make([]string, len(info.Values))
		for i, v := range info.Values {
			if len(info.CustomNames) > i && info.CustomNames[i] != "" {
				names[i] = info.CustomNames[i]
			} else {
				names[i] = converter.ToEnumValueName(v, info.BaseType)
			}
		}
		info.SanitizedNames = deduplicateNames(names)
	}
}

// deduplicateNames resolves within-list name collisions by appending numeric suffixes.
// If a name appears more than once, all occurrences get suffixed: Name0, Name1, etc.
func deduplicateNames(names []string) []string {
	// Count occurrences
	counts := make(map[string]int)
	for _, n := range names {
		counts[n]++
	}

	// Assign suffixes for duplicates
	result := make([]string, len(names))
	nextSuffix := make(map[string]int)
	for i, n := range names {
		if counts[n] > 1 {
			idx := nextSuffix[n]
			nextSuffix[n] = idx + 1
			result[i] = fmt.Sprintf("%s%d", n, idx)
		} else {
			result[i] = n
		}
	}
	return result
}

// resolveEnumCollisions determines which enums need type-name prefixes on their constants.
// An enum gets PrefixTypeName=true when:
//  1. Cross-enum collision: two different enums would produce the same constant name.
//  2. Type-name collision: an enum constant name matches any non-enum type name in the package.
//  3. Self-collision: an enum constant name matches its own type name.
//
// After the initial pass, a second pass detects collisions between post-prefix names
// (for enums already flagged) and other enums' names. This handles cases where an enum
// value like "Enum1One" collides with Enum1's prefixed constant "Enum1One".
//
// When alwaysPrefix is true, all enums are forced to use type-name prefixes.
func resolveEnumCollisions(infos []*EnumInfo, allTypeNames map[string]bool, alwaysPrefix bool) {
	if alwaysPrefix {
		for _, info := range infos {
			info.PrefixTypeName = true
		}
		return
	}

	// Pass 1: detect collisions using unprefixed names
	constOwners := make(map[string][]int)
	for i, info := range infos {
		for _, name := range info.SanitizedNames {
			constOwners[name] = append(constOwners[name], i)
		}
	}

	for i, info := range infos {
		if info.PrefixTypeName {
			continue
		}

		for _, name := range info.SanitizedNames {
			// Rule 1: Cross-enum collision
			owners := constOwners[name]
			if len(owners) > 1 {
				for _, idx := range owners {
					infos[idx].PrefixTypeName = true
				}
			}

			// Rule 2: Constant name matches a non-enum type name
			if allTypeNames[name] {
				infos[i].PrefixTypeName = true
			}

			// Rule 3: Constant name matches own type name
			if name == info.TypeName {
				infos[i].PrefixTypeName = true
			}
		}
	}

	// Pass 2: detect collisions between final constant names (post-prefix).
	// This catches cases where a prefixed name like "Enum1One" (from Enum1 + "One")
	// collides with another enum's unprefixed name "Enum1One" (from value "Enum1One").
	// Iterate until stable (no new prefixing needed).
	for {
		changed := false
		finalNames := make(map[string][]int) // final const name -> enum indices
		for i, info := range infos {
			for j := range info.SanitizedNames {
				name := info.finalConstName(j)
				finalNames[name] = append(finalNames[name], i)
			}
		}

		for _, indices := range finalNames {
			if len(indices) > 1 {
				for _, idx := range indices {
					if !infos[idx].PrefixTypeName {
						infos[idx].PrefixTypeName = true
						changed = true
					}
				}
			}
		}

		if !changed {
			break
		}
	}
}
