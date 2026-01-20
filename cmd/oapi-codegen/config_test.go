package main

import (
	"testing"

	"github.com/jinuthankachan/oapi-codegen/v2/pkg/codegen"
	"github.com/stretchr/testify/assert"
)

func TestGenerationTargets(t *testing.T) {
	tests := []struct {
		name          string
		targets       []string
		expected      codegen.GenerateOptions
		expectedError string
	}{
		{
			name:    "echo-server",
			targets: []string{"echo-server"},
			expected: codegen.GenerateOptions{
				EchoServer: true,
			},
		},
		{
			name:    "echo5-server",
			targets: []string{"echo5-server"},
			expected: codegen.GenerateOptions{
				Echo5Server: true,
			},
		},
		{
			name:    "echo5",
			targets: []string{"echo5"},
			expected: codegen.GenerateOptions{
				Echo5Server: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &codegen.Configuration{}
			err := generationTargets(cfg, tt.targets)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, cfg.Generate)
			}
		})
	}
}
