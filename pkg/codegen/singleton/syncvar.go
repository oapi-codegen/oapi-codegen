package singleton

import (
	"fmt"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
)

// GlobalState stores all global state. Please don't put global state anywhere
// else so that we can easily track it.
var GlobalState struct {
	Options       Configuration
	Spec          *openapi3.T
	ImportMapping ImportMap
}

// GoImport represents a go package to be imported in the generated code
type GoImport struct {
	Name string // package name
	Path string // package path
}

// String returns a go import statement
func (gi GoImport) String() string {
	if gi.Name != "" {
		return fmt.Sprintf("%s %q", gi.Name, gi.Path)
	}
	return fmt.Sprintf("%q", gi.Path)
}

// ImportMap maps external OpenAPI specifications files/urls to external go packages
type ImportMap map[string]GoImport

// ImportMappingCurrentPackage allows an Import Mapping to map to the current package, rather than an external package.
// This allows users to split their OpenAPI specification across multiple files, but keep them in the same package, which can reduce a bit of the overhead for users.
// We use `-` to indicate that this is a bit of a special case
const ImportMappingCurrentPackage = "-"

// GoImports returns a slice of go import statements
func (im ImportMap) GoImports() []string {
	goImports := make([]string, 0, len(im))
	for _, v := range im {
		if v.Path == ImportMappingCurrentPackage {
			continue
		}
		goImports = append(goImports, v.String())
	}
	return goImports
}

func ConstructImportMapping(importMapping map[string]string) ImportMap {
	var (
		pathToName = map[string]string{}
		result     = ImportMap{}
	)

	{
		var packagePaths []string
		for _, packageName := range importMapping {
			packagePaths = append(packagePaths, packageName)
		}
		sort.Strings(packagePaths)

		for _, packagePath := range packagePaths {
			if _, ok := pathToName[packagePath]; !ok && packagePath != ImportMappingCurrentPackage {
				pathToName[packagePath] = fmt.Sprintf("externalRef%d", len(pathToName))
			}
		}
	}
	for specPath, packagePath := range importMapping {
		result[specPath] = GoImport{Name: pathToName[packagePath], Path: packagePath}
	}
	return result
}
