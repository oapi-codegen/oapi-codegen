package util

import "strings"

func IsMediaTypeJson(mediaType string) bool {
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}
