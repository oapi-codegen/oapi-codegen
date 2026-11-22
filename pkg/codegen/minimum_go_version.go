package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

const maximumDepthToSearchForGoMod = 5

// minimumGoVersionForGenerateStdHTTPServer indicates the Go 1.x minor version that the module the std-http-server is being generated into needs.
// If the version is lower, a warning should be logged.
const minimumGoVersionForGenerateStdHTTPServer = 22

func findAndParseGoModuleForDepth(dir string, maxDepth int) (string, *modfile.File, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to determine absolute path for %v: %w", dir, err)
	}
	currentDir := absDir

	for i := 0; i <= maxDepth; i++ {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			goModContent, err := os.ReadFile(goModPath)
			if err != nil {
				return "", nil, fmt.Errorf("failed to read `go.mod`: %w", err)
			}

			mod, err := modfile.ParseLax("go.mod", goModContent, nil)
			if err != nil {
				return "", nil, fmt.Errorf("failed to parse `go.mod`: %w", err)
			}

			return goModPath, mod, nil
		}

		goModPath = filepath.Join(currentDir, "tools.mod")
		if _, err := os.Stat(goModPath); err == nil {
			goModContent, err := os.ReadFile(goModPath)
			if err != nil {
				return "", nil, fmt.Errorf("failed to read `tools.mod`: %w", err)
			}

			parsedModFile, err := modfile.ParseLax("tools.mod", goModContent, nil)
			if err != nil {
				return "", nil, fmt.Errorf("failed to parse `tools.mod`: %w", err)
			}

			return goModPath, parsedModFile, nil
		}

		parentDir := filepath.Dir(currentDir)
		// NOTE that this may not work particularly well on Windows
		if parentDir == "/" {
			break
		}

		currentDir = parentDir
	}

	return "", nil, fmt.Errorf("no `go.mod` or `tools.mod` file found within %d levels upwards from %s", maxDepth, absDir)
}

// hasMinimalMinorGoDirective indicates that the Go module (`mod`) has a minor version greater than or equal to the `expected`'s
// This only applies to the `go` directive:
//
//	go 1.23
//	go 1.22.1
func hasMinimalMinorGoDirective(expected int, mod *modfile.File) bool {
	// a `go.mod` or `tools.mod` is valid without a `go` directive, which the Go toolchain treats as an old language version
	if mod == nil || mod.Go == nil {
		return false
	}

	parts := strings.Split(mod.Go.Version, ".")

	if len(parts) < 2 {
		return false
	}

	// the minor version may be followed by a pre-release suffix, i.e. `go 1.22rc1` or `go 1.23beta1`, so only read its leading digits
	minor := parts[1]
	if end := strings.IndexFunc(minor, func(r rune) bool { return r < '0' || r > '9' }); end != -1 {
		minor = minor[:end]
	}

	actual, err := strconv.Atoi(minor)
	if err != nil {
		return false
	}

	if actual < expected {
		return false
	}

	return true
}
