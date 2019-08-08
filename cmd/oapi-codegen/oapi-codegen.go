// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/weberr13/oapi-codegen/pkg/codegen"
	"github.com/weberr13/oapi-codegen/pkg/util"
)

func errExit(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func usageErr(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	flag.PrintDefaults()
	os.Exit(1)
}

// TODO define a type to allow recording of Type -> Import Path as args
// TODO this is temporary until we sort the config question
type refImports []string

func (ri *refImports) String() string {
	return "Reference Import arg"
}

func (ri *refImports) Set(r string) error {
	*ri = append(*ri, r)
	return nil
}

// TODO define a type to allow recording of package -> Import Path as args
// TODO this is temporary until we sort the config question
type pkgImports []string

func (pi *pkgImports) String() string {
	return "Pkg Import arg"
}

func (pi *pkgImports) Set(r string) error {
	*pi = append(*pi, r)
	return nil
}

func main() {
	var (
		packageName string
		generate    string
		outputFile  string
		includeTags string
		excludeTags string
		refImports  refImports
		pkgImports  pkgImports
		allowRefs   bool
		insecure    bool
	)

	flag.StringVar(&packageName, "package", "", "The package name for generated code")
	flag.StringVar(&generate, "generate", "types,client,server,spec",
		`Comma-separated list of code to generate; valid options: "types", "client", "chi-server", "server", "skip-fmt", "spec"`)
	flag.StringVar(&outputFile, "o", "", "Where to output generated code, stdout is default")
	flag.StringVar(&includeTags, "include-tags", "", "Only include operations with the given tags. Comma-separated list of tags.")
	flag.Var(&refImports, "ri", "Repeated reference import statements of the form Type=<[package:]import path>")
	flag.Var(&pkgImports, "pi", "Repeated package import statements of the form <package>=<[package:]import path>")
	flag.BoolVar(&allowRefs, "extrefs", false, "Allow resolving external references")
	flag.BoolVar(&insecure, "insecure", false, "Allow resolving remote URL's that have bad SSL/TLS")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please specify a path to a OpenAPI 3.0 spec file")
		os.Exit(1)
	}

	// If the package name has not been specified, we will use the name of the
	// swagger file.
	if packageName == "" {
		path := flag.Arg(0)
		baseName := filepath.Base(path)
		// Split the base name on '.' to get the first part of the file.
		nameParts := strings.Split(baseName, ".")
		packageName = codegen.ToCamelCase(nameParts[0])
	}

	opts := codegen.Options{}
	for _, g := range splitCSVArg(generate) {
		switch g {
		case "client":
			opts.GenerateClient = true
		case "chi-server":
			opts.GenerateChiServer = true
		case "server":
			opts.GenerateEchoServer = true
		case "types":
			opts.GenerateTypes = true
		case "spec":
			opts.EmbedSpec = true
		case "skip-fmt":
			opts.SkipFmt = true
		case "resolved-spec":
			opts.ClearRefsSpec = true
			opts.EmbedSpec = true
		default:
			// never returns
			usageErr("unknown generate option %s\n", g)
		}
	}

	opts.IncludeTags = splitCSVArg(includeTags)
	opts.ExcludeTags = splitCSVArg(excludeTags)

	if opts.GenerateEchoServer && opts.GenerateChiServer {
		errExit("can not specify both server and chi-server targets simultaneously")
	}

	if opts.ClearRefsSpec && (opts.GenerateClient || opts.GenerateServer || opts.GenerateTypes) {
		// never returns
		usageErr("resolved-spec option is only valid when specified on its own")
	}

	// Add user defined type -> import mapping
	opts.ImportedTypes = importPackages(pkgImports, allowRefs, insecure)
	// override with specific imports if so indicated
	for k, v := range importTypes(refImports) {
		opts.ImportedTypes[k] = v
	}

	swagger, err := util.LoadSwagger(flag.Arg(0), allowRefs, insecure, opts.ClearRefsSpec)
	if err != nil {
		errExit("error loading swagger spec:\n%s\n", err)
	}

	code, err := codegen.Generate(swagger, packageName, opts)
	if err != nil {
		errExit("error generating code: %s\n", err)
	}

	if outputFile != "" {
		err = ioutil.WriteFile(outputFile, []byte(code), 0644)
		if err != nil {
			errExit("error writing generated code to file: %s\n", err)
		}
	} else {
		fmt.Println(code)
	}
}

func splitCSVArg(input string) []string {
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

func mapArgSlice(as []string, an string) map[string]string {
	m := map[string]string{}
	for _, ri := range as {
		parts := strings.Split(ri, "=")
		if len(parts) != 2 {
			fmt.Printf("invalid %s arg. %s\n", an, ri)
			flag.PrintDefaults()
			os.Exit(1)
		}
		m[parts[0]] = parts[1]
	}
	return m
}

func importPackages(imports pkgImports, allowRefs, insecure bool) map[string]codegen.TypeImportSpec {
	importedTypes := map[string]codegen.TypeImportSpec{}
	pi := mapArgSlice(imports, "package import")

	for sr, p := range pi {
		var u *url.URL
		var err error

		if u, err = url.Parse(sr); err != nil {
			errExit("package import: specified schema ref is not a URL:\n%s\n", err)
		}
		swagger, err := util.LoadSwaggerFromURL(u, allowRefs, insecure)
		if err != nil {
			errExit("package import: error loading swagger spec:\n%s\n", err)
		}

		// iterate over model to find schema names - extract only top level names
		for n := range swagger.Components.Schemas {
			if _, ok := importedTypes[n]; !ok {
				importedTypes[n] = getImportedType(n, p)
			}
		}
	}
	return importedTypes
}

func importTypes(typeImports refImports) map[string]codegen.TypeImportSpec {
	importedTypes := map[string]codegen.TypeImportSpec{}
	pi := mapArgSlice(typeImports, "type imports")

	for t, p := range pi {
		importedTypes[t] = getImportedType(t, p)
	}
	return importedTypes
}

func getImportedType(typeName, pkgImport string) codegen.TypeImportSpec {
	var pkgName string
	var impPath string

	// if a package name was specified in the form pkg:<import path>, use that, otherwise use last part of import path
	parts := strings.Split(pkgImport, ":")
	if len(parts) > 2 {
		errExit("Parsing type import: too many fragments. At most one ':' expected (type:%s, import:%s)", typeName, pkgImport)
	}

	if len(parts) == 2 {
		pkgName = parts[0]
		impPath = parts[1]
	} else {
		parts = strings.Split(pkgImport, "/")
		pkgName = parts[len(parts)-1]
		impPath = pkgImport
	}

	return codegen.NewTypeImportSpec(typeName, pkgName, impPath)
}
