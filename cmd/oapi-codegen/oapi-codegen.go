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

// Generate is a flag for setting what code to generate.
type Generate string

const (
	// All generates all code and is the default.
	All Generate = "all"
	// Server generates the types and server implementation.
	Server Generate = "server"
	// Client generates the types and the client implementation.
	Client Generate = "client"
	// Types generates the request, responses, and parameters structs.
	Types Generate = "types"
)

func (g Generate) String() string {
	if g == "" {
		return string(All)
	}
	return string(g)
}

// Set checks for valid Generate options. Uses All by default.
func (g Generate) Set(s string) error {
	// default generate all
	if s == "" {
		g = All
		return nil
	}

	g = Generate(s)
	switch g {
	case All, Server, Client, Types:
		return nil
	default:
		return fmt.Errorf(`must be "all", "server", "client", or "types"`)
	}
}

func main() {
	var packageName string
	var generate Generate
	flag.StringVar(&packageName, "package", "", "The package name for generated code")
	flag.Var(&generate, "generate", "The code to generate; valid options: all (default), server, client, types")
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

	swagger, err := util.LoadSwagger(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	stubs, err := codegen.GenerateServer(swagger, packageName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating server stubs: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(stubs)
}
