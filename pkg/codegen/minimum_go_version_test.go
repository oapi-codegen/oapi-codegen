package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
)

func TestMinimumGoVersion_hasMinimalMinorGoDirective(t *testing.T) {
	tests := []struct {
		name  string
		gomod string
		want  bool
	}{
		{name: "no go directive", gomod: "module example.com/m\n", want: false},
		{name: "older minor", gomod: "module example.com/m\n\ngo 1.21\n", want: false},
		{name: "older minor with patch", gomod: "module example.com/m\n\ngo 1.21.0\n", want: false},
		{name: "exact minor", gomod: "module example.com/m\n\ngo 1.22\n", want: true},
		{name: "newer minor with patch", gomod: "module example.com/m\n\ngo 1.23.1\n", want: true},
		{name: "release candidate of the minor", gomod: "module example.com/m\n\ngo 1.22rc1\n", want: true},
		{name: "release candidate of an older minor", gomod: "module example.com/m\n\ngo 1.21rc2\n", want: false},
		{name: "beta of the minor with patch", gomod: "module example.com/m\n\ngo 1.22.0beta1\n", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod, err := modfile.ParseLax("go.mod", []byte(tt.gomod), nil)
			require.NoError(t, err)

			assert.Equal(t, tt.want, hasMinimalMinorGoDirective(minimumGoVersionForGenerateStdHTTPServer, mod))
		})
	}
}

func TestMinimumGoVersion_warningForStdHTTPWithoutGoDirective(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/m\n"), 0o600))
	t.Chdir(dir)

	var warning string
	require.NotPanics(t, func() {
		warning = GenerateOptions{StdHTTPServer: true}.warningForStdHTTP()
	})
	assert.Contains(t, warning, "needs to be bumped")
}
