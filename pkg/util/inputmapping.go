package util

import (
	"fmt"
	"strings"
)

// The input mapping is experessed on the command line as `key1:value1,key2:value2,...`
// We parse it here, but need to keep in mind that keys or values may contain
// commas and colons. We will allow escaping those using double quotes, so
// when passing in "key1":"value1", we will not look inside the quoted sections.
func ParseCommandlineMap(src string) (map[string]string, error) {
	result := make(map[string]string)
	tuples := splitString(src, ',')
	for _, t := range tuples {
		kv := splitString(t, ':')
		if len(kv) != 2 {
			return nil, fmt.Errorf("expected key:value, got :%s", t)
		}
		key := strings.TrimLeft(kv[0], `"`)
		key = strings.TrimRight(key, `"`)

		value := strings.TrimLeft(kv[1], `"`)
		value = strings.TrimRight(value, `"`)

		result[key] = value
	}
	return result, nil
}

// ParseCommandLineList parses comma separated string lists which are passed
// in on the command line. Spaces are trimmed off both sides of result
// strings.
func ParseCommandLineList(input string) []string {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return nil
	}
	splitInput := strings.Split(input, ",")
	args := make([]string, 0, len(splitInput))
	for _, s := range splitInput {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			args = append(args, s)
		}
	}
	return args
}

// This function splits a string along the specifed separator, but it
// ignores anything between double quotes for splitting. We do simple
// inside/outside quote counting. Quotes are not stripped from output.
func splitString(s string, sep rune) []string {
	const escapeChar rune = '"'

	var parts []string
	var part string
	inQuotes := false

	for _, c := range s {
		if c == escapeChar {
			if inQuotes {
				inQuotes = false
			} else {
				inQuotes = true
			}
		}

		// If we've gotten the separator rune, consider the previous part
		// complete, but only if we're outside of quoted sections
		if c == sep && !inQuotes {
			parts = append(parts, part)
			part = ""
			continue
		}
		part = part + string(c)
	}
	return append(parts, part)
}
