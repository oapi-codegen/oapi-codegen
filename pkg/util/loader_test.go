package util

import (
	"testing"
)

func TestLoadSwagger(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{"#1 file", "../../examples/petstore-expanded/petstore-expanded.yaml", false},
		{"#2 file", "file://../../examples/petstore-expanded/petstore-expanded.yaml", false},
		{"#3 file", "./../../examples/petstore-expanded/petstore-expanded.yaml", false},
		{"#4 file", "file://./../../examples/petstore-expanded/petstore-expanded.yaml", false},

		{"#5 https", "https://raw.githubusercontent.com/deepmap/oapi-codegen/master/examples/petstore-expanded/petstore-expanded.yaml", false},

		{"#6 git/https", "https://github.com/deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml", false},

		{"#7 git/https", "https://github.com/deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml@master", false},

		// git repos (github flavor)
		// these tests require a configured ssh client and a private/public
		// key pair of which the public one is known by github
		/*
			{"#8 git", "git@github.com:22:deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml", false},
			{"#9 git", "git@github.com:22:deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml@master", false},

			{"#10 git", "git@github.com:deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml", false},
			{"#11 git", "git@github.com:deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml@master", false},

			{"#12 git", "ssh://git@github.com:deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml", false},
			{"#13 git", "ssh://git@github.com:deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml@master", false},

			{"#14 git", "ssh://git@github.com:22:deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml", false},
			{"#15 git", "ssh://git@github.com:22:deepmap/oapi-codegen.git/examples/petstore-expanded/petstore-expanded.yaml@master", false},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swagger, err := LoadSwagger(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSwagger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if swagger == nil {
				t.Errorf("LoadSwagger() returned nil object")
				return
			}
		})
	}
}
