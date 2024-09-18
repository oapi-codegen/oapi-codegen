package util

import (
	"testing"
)

func TestIsMediaTypeJson(t *testing.T) {
	type test struct {
		name      string
		mediaType string
		want      bool
	}

	suite := []test{
		{
			name: "When no MediaType, returns false",
			want: false,
		},
		{
			name:      "When not a JSON MediaType, returns false",
			mediaType: "application/pdf",
			want:      false,
		},
		{
			name:      "When MediaType ends with json, but isn't JSON, returns false",
			mediaType: "application/notjson",
			want:      false,
		},
		{
			name:      "When MediaType is application/json, returns true",
			mediaType: "application/json",
			want:      true,
		},
		{
			name:      "When MediaType is application/json-patch+json, returns true",
			mediaType: "application/json-patch+json",
			want:      true,
		},
		{
			name:      "When MediaType is application/vnd.api+json, returns true",
			mediaType: "application/vnd.api+json",
			want:      true,
		},
		{
			// NOTE that this _technically_ isn't a standard extension to JSON https://www.iana.org/assignments/media-types/application/json but due to the fact that several APIs do use it, we should support it
			name:      "When MediaType is application/json;v=1, returns true",
			mediaType: "application/json;v=1",
			want:      true,
		},
		{
			// NOTE that this _technically_ isn't a standard extension to JSON https://www.iana.org/assignments/media-types/application/json but due to the fact that several APIs do use it, we should support it
			name:      "When MediaType is application/json;version=1, returns true",
			mediaType: "application/json;version=1",
			want:      true,
		},
	}
	for _, test := range suite {
		t.Run(test.name, func(t *testing.T) {
			got := IsMediaTypeJson(test.mediaType)

			if got != test.want {
				t.Fatalf("IsJson validation failed. Want [%v] Got [%v]", test.want, got)
			}
		})
	}
}
