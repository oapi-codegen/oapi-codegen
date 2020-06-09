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
	"path"
	"path/filepath"
	"strings"

	"github.com/tidepool-org/oapi-codegen/pkg/codegen"
	"github.com/tidepool-org/oapi-codegen/pkg/util"
)

func errExit(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {
	var (
		packageName  string
		generate     string
		outputFile   string
		includeTags  string
		excludeTags  string
		templatesDir string
	)
	flag.StringVar(&packageName, "package", "", "The package name for generated code")
	flag.StringVar(&generate, "generate", "types,client,server,spec",
		`Comma-separated list of code to generate; valid options: "types", "client", "chi-server", "server", "spec", "skip-fmt", "skip-prune"`)
	flag.StringVar(&outputFile, "o", "", "Where to output generated code, stdout is default")
	flag.StringVar(&includeTags, "include-tags", "", "Only include operations with the given tags. Comma-separated list of tags.")
	flag.StringVar(&excludeTags, "exclude-tags", "", "Exclude operations that are tagged with the given tags. Comma-separated list of tags.")
	flag.StringVar(&templatesDir, "templates", "", "Path to directory containing user templates")
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
		case "skip-prune":
			opts.SkipPrune = true
		default:
			fmt.Printf("unknown generate option %s\n", g)
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	opts.IncludeTags = splitCSVArg(includeTags)
	opts.ExcludeTags = splitCSVArg(excludeTags)

	if opts.GenerateEchoServer && opts.GenerateChiServer {
		errExit("can not specify both server and chi-server targets simultaneously")
	}

	swagger, err := util.LoadSwagger(flag.Arg(0))
	if err != nil {
		errExit("error loading swagger spec\n: %s", err)
	}

	templates, err := loadTemplateOverrides(templatesDir)
	if err != nil {
		errExit("error loading template overrides: %s\n", err)
	}
	opts.UserTemplates = templates

	code, err := codegen.Generate(swagger, packageName, opts)
	if err != nil {
		errExit("error generating code: %s\n", err)
	}

	if outputFile != "" {
		err = ioutil.WriteFile(outputFile, []byte(code), 0644)
		if err != nil {
			errExit("error writing generated code to file: %s", err)
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

func loadTemplateOverrides(templatesDir string) (map[string]string, error) {
	var templates = make(map[string]string)

	if templatesDir == "" {
		return templates, nil
	}

	files, err := ioutil.ReadDir(templatesDir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		data, err := ioutil.ReadFile(path.Join(templatesDir, f.Name()))
		if err != nil {
			return nil, err
		}
		templates[f.Name()] = string(data)
	}

	return templates, nil
}
