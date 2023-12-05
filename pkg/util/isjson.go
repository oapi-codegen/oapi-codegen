package util

import "strings"

func IsMediaTypeJson(mediaType string) bool {
	return strings.HasPrefix(mediaType, "application/json") || strings.HasSuffix(mediaType, "+json")
}
