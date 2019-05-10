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

// Generate are possible options to the generate flag.
type Generate string

const (
	// Server generates the types and server implementation.
	Server Generate = "server"
	// Client generates the types and the client implementation.
	Client Generate = "client"
)

// Generators is a flag for setting what code to generate. Acts like
// a unique set of options.
type Generators map[Generate]struct{}

func (g Generators) String() string {
	opts := []string{}
	for k := range g {
		opts = append(opts, string(k))
	}
	return strings.Join(opts, ",")
}

// Set checks for valid Generate options. Uses All by default.
func (g Generators) Set(value string) error {
	// by default generate both the server and the client
	if value == "" {
		g[Server] = struct{}{}
		g[Client] = struct{}{}
		return nil
	}

	for _, s := range strings.Split(value, ",") {
		gen := Generate(s)
		// default generate all
		if gen == "" {
			g[Server] = struct{}{}
			g[Client] = struct{}{}
		}

		switch gen {
		case Server, Client:
			g[gen] = struct{}{}
		default:
			return fmt.Errorf(`must be "client" or "server"`)
		}
	}

	return nil
}

func main() {
	var packageName string
	generators := make(Generators)
	flag.StringVar(&packageName, "package", "", "The package name for generated code")
	flag.Var(generators, "generate", `Comma-separated list of code to generate; valid options: "client","server"  (default client,server) `)
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
