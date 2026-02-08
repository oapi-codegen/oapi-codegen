package codegen

import (
	"strings"
	"unicode"
)

// NameMangling configures how OpenAPI names are converted to valid Go identifiers.
type NameMangling struct {
	// CharacterSubstitutions maps characters to their word replacements.
	// Used when these characters appear at the start of a name.
	// Example: '$' -> "DollarSign", '-' -> "Minus"
	CharacterSubstitutions map[string]string `yaml:"character-substitutions,omitempty"`

	// WordSeparators is a string of characters that mark word boundaries.
	// When encountered, the next letter is capitalized.
	// Example: "-_. " means "foo-bar" becomes "FooBar"
	WordSeparators string `yaml:"word-separators,omitempty"`

	// NumericPrefix is prepended when a name starts with a digit.
	// Example: "N" means "123foo" becomes "N123foo"
	NumericPrefix string `yaml:"numeric-prefix,omitempty"`

	// KeywordPrefix is prepended when a name conflicts with a Go keyword.
	// Example: "_" means "type" becomes "_type"
	KeywordPrefix string `yaml:"keyword-prefix,omitempty"`

	// Initialisms is a list of words that should be all-uppercase.
	// Example: ["ID", "HTTP", "URL"] means "userId" becomes "UserID"
	Initialisms []string `yaml:"initialisms,omitempty"`
}

// DefaultNameMangling returns sensible defaults for name mangling.
func DefaultNameMangling() NameMangling {
	return NameMangling{
		CharacterSubstitutions: map[string]string{
			"$":  "DollarSign",
			"-":  "Minus",
			"+":  "Plus",
			"&":  "And",
			"|":  "Or",
			"~":  "Tilde",
			"=":  "Equal",
			">":  "GreaterThan",
			"<":  "LessThan",
			"#":  "Hash",
			".":  "Dot",
			"*":  "Asterisk",
			"^":  "Caret",
			"%":  "Percent",
			"_":  "Underscore",
			"@":  "At",
			"!":  "Bang",
			"?":  "Question",
			"/":  "Slash",
			"\\": "Backslash",
			":":  "Colon",
			";":  "Semicolon",
			"'":  "Apos",
			"\"": "Quote",
			"`":  "Backtick",
			"(":  "LParen",
			")":  "RParen",
			"[":  "LBracket",
			"]":  "RBracket",
			"{":  "LBrace",
			"}":  "RBrace",
		},
		WordSeparators: "-#@!$&=.+:;_~ (){}[]|<>?/\\",
		NumericPrefix:  "N",
		KeywordPrefix:  "_",
		Initialisms: []string{
			"ACL", "API", "ASCII", "CPU", "CSS", "DB", "DNS", "EOF",
			"GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON",
			"QPS", "RAM", "RPC", "SLA", "SMTP", "SQL", "SSH", "TCP",
			"TLS", "TTL", "UDP", "UI", "UID", "GID", "URI", "URL",
			"UTF8", "UUID", "VM", "XML", "XMPP", "XSRF", "XSS",
			"SIP", "RTP", "AMQP", "TS",
		},
	}
}

// Merge returns a new NameMangling with user values overlaid on defaults.
// Non-zero user values override defaults.
func (n NameMangling) Merge(user NameMangling) NameMangling {
	result := n

	// Merge character substitutions (user overrides/adds to defaults)
	if len(user.CharacterSubstitutions) > 0 {
		merged := make(map[string]string, len(n.CharacterSubstitutions))
		for k, v := range n.CharacterSubstitutions {
			merged[k] = v
		}
		for k, v := range user.CharacterSubstitutions {
			if v == "" {
				// Empty string means "remove this substitution"
				delete(merged, k)
			} else {
				merged[k] = v
			}
		}
		result.CharacterSubstitutions = merged
	}

	if user.WordSeparators != "" {
		result.WordSeparators = user.WordSeparators
	}
	if user.NumericPrefix != "" {
		result.NumericPrefix = user.NumericPrefix
	}
	if user.KeywordPrefix != "" {
		result.KeywordPrefix = user.KeywordPrefix
	}
	if len(user.Initialisms) > 0 {
		result.Initialisms = user.Initialisms
	}

	return result
}

// NameSubstitutions holds direct name overrides for generated identifiers.
type NameSubstitutions struct {
	// TypeNames maps generated type names to user-preferred names.
	// Example: {"MyGeneratedType": "MyPreferredName"}
	TypeNames map[string]string `yaml:"type-names,omitempty"`

	// PropertyNames maps generated property/field names to user-preferred names.
	// Example: {"GeneratedField": "PreferredField"}
	PropertyNames map[string]string `yaml:"property-names,omitempty"`
}

// NameConverter handles converting OpenAPI names to Go identifiers.
type NameConverter struct {
	mangling      NameMangling
	substitutions NameSubstitutions
	initialismSet map[string]bool
}

// NewNameConverter creates a NameConverter with the given configuration.
func NewNameConverter(mangling NameMangling, substitutions NameSubstitutions) *NameConverter {
	initialismSet := make(map[string]bool, len(mangling.Initialisms))
	for _, init := range mangling.Initialisms {
		initialismSet[strings.ToUpper(init)] = true
	}
	return &NameConverter{
		mangling:      mangling,
		substitutions: substitutions,
		initialismSet: initialismSet,
	}
}

// ToTypeName converts an OpenAPI schema name to a Go type name.
func (c *NameConverter) ToTypeName(name string) string {
	// Check for direct substitution first
	if sub, ok := c.substitutions.TypeNames[name]; ok {
		return sub
	}
	return c.toGoIdentifier(name, true)
}

// ToTypeNamePart converts a name to a type name component that will be joined with others.
// Unlike ToTypeName, it doesn't add a numeric prefix since the result won't be the start of an identifier.
func (c *NameConverter) ToTypeNamePart(name string) string {
	// Check for direct substitution first
	if sub, ok := c.substitutions.TypeNames[name]; ok {
		return sub
	}
	return c.toGoIdentifierPart(name)
}

// ToPropertyName converts an OpenAPI property name to a Go field name.
func (c *NameConverter) ToPropertyName(name string) string {
	// Check for direct substitution first
	if sub, ok := c.substitutions.PropertyNames[name]; ok {
		return sub
	}
	return c.toGoIdentifier(name, true)
}

// ToEnumValueName converts a raw enum value to a valid Go identifier.
// For integer enums (baseType is int, int32, int64, etc.), numeric values
// get the configured NumericPrefix, negative values get "Minus" prefix.
// For string enums, the value is processed through the full NameConverter pipeline.
func (c *NameConverter) ToEnumValueName(value string, baseType string) string {
	if value == "" {
		return "Empty"
	}

	// Handle integer enum values
	if isIntegerType(baseType) && len(value) > 0 {
		first := value[0]
		if first >= '0' && first <= '9' {
			return c.mangling.NumericPrefix + value
		}
		if first == '-' && len(value) > 1 {
			return "Minus" + value[1:]
		}
	}

	return c.toGoIdentifier(value, true)
}

// isIntegerType returns true if the Go type is an integer type.
func isIntegerType(baseType string) bool {
	switch baseType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return true
	default:
		return false
	}
}

// ToVariableName converts an OpenAPI name to a Go variable name (unexported).
func (c *NameConverter) ToVariableName(name string) string {
	id := c.toGoIdentifier(name, false)
	if id == "" {
		return id
	}
	// Make first letter lowercase
	runes := []rune(id)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// toGoIdentifier converts a name to a valid Go identifier.
func (c *NameConverter) toGoIdentifier(name string, exported bool) string {
	if name == "" {
		return "Empty"
	}

	// Build the identifier with prefix handling
	var result strings.Builder
	prefix := c.getPrefix(name)
	result.WriteString(prefix)

	// Convert the rest using word boundaries
	capitalizeNext := exported || prefix != ""
	prevWasDigit := false
	for _, r := range name {
		if c.isWordSeparator(r) {
			capitalizeNext = true
			prevWasDigit = false
			continue
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			// Skip invalid characters (already handled by prefix if at start)
			capitalizeNext = true
			prevWasDigit = false
			continue
		}

		// Capitalize after digits
		if prevWasDigit && unicode.IsLetter(r) {
			capitalizeNext = true
		}

		if capitalizeNext && unicode.IsLetter(r) {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}

		prevWasDigit = unicode.IsDigit(r)
	}

	id := result.String()
	if id == "" {
		return "Empty"
	}

	// Apply initialism fixes
	id = c.applyInitialisms(id)

	return id
}

// toGoIdentifierPart converts a name to a Go identifier component (for joining with others).
// It doesn't add a numeric prefix since the result won't necessarily be at the start of an identifier.
func (c *NameConverter) toGoIdentifierPart(name string) string {
	if name == "" {
		return ""
	}

	// Build the identifier without numeric prefix (but still handle special characters at start)
	var result strings.Builder

	// Only add prefix for non-digit special characters at the start
	firstRune := []rune(name)[0]
	if !unicode.IsLetter(firstRune) && !unicode.IsDigit(firstRune) {
		firstChar := string(firstRune)
		if sub, ok := c.mangling.CharacterSubstitutions[firstChar]; ok {
			result.WriteString(sub)
		} else {
			result.WriteString("X")
		}
	}

	// Convert the rest using word boundaries (always capitalize since this is a part)
	capitalizeNext := true
	prevWasDigit := false
	for _, r := range name {
		if c.isWordSeparator(r) {
			capitalizeNext = true
			prevWasDigit = false
			continue
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			// Skip invalid characters (already handled by prefix if at start)
			capitalizeNext = true
			prevWasDigit = false
			continue
		}

		// Capitalize after digits
		if prevWasDigit && unicode.IsLetter(r) {
			capitalizeNext = true
		}

		if capitalizeNext && unicode.IsLetter(r) {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}

		prevWasDigit = unicode.IsDigit(r)
	}

	id := result.String()

	// Apply initialism fixes
	id = c.applyInitialisms(id)

	return id
}

// getPrefix returns the prefix needed for names starting with invalid characters.
func (c *NameConverter) getPrefix(name string) string {
	if name == "" {
		return ""
	}

	firstRune := []rune(name)[0]

	// Check if starts with digit
	if unicode.IsDigit(firstRune) {
		return c.mangling.NumericPrefix
	}

	// Check if starts with letter (valid, no prefix needed)
	if unicode.IsLetter(firstRune) {
		return ""
	}

	// Check character substitutions
	firstChar := string(firstRune)
	if sub, ok := c.mangling.CharacterSubstitutions[firstChar]; ok {
		return sub
	}

	// Unknown special character, use generic prefix
	return "X"
}

// isWordSeparator returns true if the rune is a word separator.
func (c *NameConverter) isWordSeparator(r rune) bool {
	return strings.ContainsRune(c.mangling.WordSeparators, r)
}

// applyInitialisms uppercases known initialisms in the identifier.
// It detects initialisms at word boundaries in PascalCase identifiers.
func (c *NameConverter) applyInitialisms(name string) string {
	if len(name) == 0 {
		return name
	}

	// Split the identifier into "words" based on case transitions
	// e.g., "UserId" -> ["User", "Id"], "HTTPUrl" -> ["HTTP", "Url"]
	words := splitPascalCase(name)

	// Check each word against initialisms
	for i, word := range words {
		upper := strings.ToUpper(word)
		if c.initialismSet[upper] {
			words[i] = upper
		}
	}

	return strings.Join(words, "")
}

// splitPascalCase splits a PascalCase identifier into words.
// e.g., "UserId" -> ["User", "Id"], "HTTPServer" -> ["HTTP", "Server"]
func splitPascalCase(s string) []string {
	if len(s) == 0 {
		return nil
	}

	var words []string
	var currentWord strings.Builder

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if i == 0 {
			currentWord.WriteRune(r)
			continue
		}

		prevUpper := unicode.IsUpper(runes[i-1])
		currUpper := unicode.IsUpper(r)
		currDigit := unicode.IsDigit(r)

		// Start new word on:
		// 1. Lowercase to uppercase transition (e.g., "userId" -> "user" | "Id")
		// 2. Multiple uppercase followed by lowercase (e.g., "HTTPServer" -> "HTTP" | "Server")
		if currUpper && !prevUpper {
			// Lowercase to uppercase: start new word
			words = append(words, currentWord.String())
			currentWord.Reset()
			currentWord.WriteRune(r)
		} else if currUpper && prevUpper && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
			// Uppercase followed by lowercase, and previous was uppercase
			// This is the start of a new word after an acronym
			// e.g., in "HTTPServer", 'S' starts a new word
			words = append(words, currentWord.String())
			currentWord.Reset()
			currentWord.WriteRune(r)
		} else if currDigit && !unicode.IsDigit(runes[i-1]) {
			// Transition to digit: start new word
			words = append(words, currentWord.String())
			currentWord.Reset()
			currentWord.WriteRune(r)
		} else if !currDigit && unicode.IsDigit(runes[i-1]) {
			// Transition from digit: start new word
			words = append(words, currentWord.String())
			currentWord.Reset()
			currentWord.WriteRune(r)
		} else {
			currentWord.WriteRune(r)
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}
