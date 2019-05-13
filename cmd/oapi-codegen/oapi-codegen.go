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
	"os"
	"path/filepath"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/util"
)

func main() {
	var (
		packageName string
		generate    string
	)
	flag.StringVar(&packageName, "package", "", "The package name for generated code")
	flag.StringVar(&generate, "generate", "client,server", `Comma-separated list of code to generate; valid options: "client", "server"  (default client,server)`)
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
		default:
			fmt.Printf("unknown generate option %s\n", g)
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	swagger, err := util.LoadSwagger(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	code, err := codegen.Generate(swagger, packageName, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(code)
}
