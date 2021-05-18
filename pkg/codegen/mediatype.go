package codegen

import (
	"regexp"
	"strings"
)

const (
	DefaultVendorJSONRegex = `^application/vnd.(?P<tag>[a-zA-Z0-9-.]+)\+json$`
)

var vendorJSONRegex *regexp.Regexp

func InitVendorJSONRegex(pattern string) {
	vendorJSONRegex = regexp.MustCompile(pattern)
}

func isJSON(mediaType string) bool {
	return strings.EqualFold(strings.ToLower(mediaType), "application/json")
}

func isVendorJSON(mediaType string) bool {
	if vendorJSONRegex == nil {
		return false
	}

	return vendorJSONRegex.MatchString(mediaType)
}

// Return a tag that is (hopefully) unique for the given media type.
// If vendorJSONRegex contains a subexpression (named "tag"), we return
// a cleaned-up version of the submatch.
// Otherwise, a tag that corresponds to the whole match is returned.
func getTagForVendorJSON(mediaType string) string {
	if vendorJSONRegex == nil {
		return ""
	}

	matches := vendorJSONRegex.FindStringSubmatch(mediaType)
	if matches == nil {
		return ""
	}

	index := vendorJSONRegex.SubexpIndex("tag")
	if index == -1 {
		// vendorJSONRegex doesn't have the subexpression "tag"
		return ToCamelCase(matches[0])
	}

	return ToCamelCase(matches[index])
}

// Return a type name for the given schema name, media type pair.
// This generates a unique type name as long as the submatch by the vendor JSON regex
// returns a unique string (e.g., a version if the media type is used for versioning
// requests/responses)
func getVendorJSONTypeName(schemaName, mediaType string) string {
	return SchemaNameToTypeName(schemaName) + getTagForVendorJSON(mediaType)
}
