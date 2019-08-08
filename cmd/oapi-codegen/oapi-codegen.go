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
	"os"
	"path/filepath"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/util"
)

func errExit(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
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

func main() {
	var (
		packageName string
		generate    string
		outputFile  string
		refImports  refImports
		allowRefs   bool
		insecure    bool
	)

	flag.StringVar(&packageName, "package", "", "The package name for generated code")
	flag.StringVar(&generate, "generate", "types,client,server,spec",
		`Comma-separated list of code to generate; valid options: "types", "client", "server", "spec"  (default types,client,server,"spec")`)
	flag.StringVar(&outputFile, "o", "", "Where to output generated code, stdout is default")
	flag.Var(&refImports, "ri", "Repeated reference import statements of the form Type=<import path>")
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
	for _, g := range strings.Split(generate, ",") {
		switch g {
		case "client":
			opts.GenerateClient = true
		case "server":
			opts.GenerateServer = true
		case "types":
			opts.GenerateTypes = true
		case "spec":
			opts.EmbedSpec = true
		default:
			fmt.Printf("unknown generate option %s\n", g)
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// Add user defined type -> import mapping
	typeImports := map[string]string{}
	for _, ri := range refImports {
		parts := strings.Split(ri, "=")
		if len(parts) != 2 {
			fmt.Printf("invalid ref import arg. %s\n", ri)
			flag.PrintDefaults()
			os.Exit(1)
		}
		typeImports[parts[0]] = parts[1]
	}
	opts.TypeImports = typeImports

	swagger, err := util.LoadSwagger(flag.Arg(0), allowRefs, insecure)
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
