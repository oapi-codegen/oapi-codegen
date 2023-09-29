package util

import "strings"

func IsMediaTypeJson(mediaType string) bool {
	return mediaType == "application/json" ||
		strings.ToLower(mediaType) == "application/json; charset=utf-8" ||
		strings.HasSuffix(mediaType, "+json")
}
