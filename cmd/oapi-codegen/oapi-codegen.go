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
	"runtime/debug"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/util"
)

func errExit(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

var (
	flagPackageName    string
	flagGenerate       string
	flagOutputFile     string
	flagIncludeTags    string
	flagExcludeTags    string
	flagTemplatesDir   string
	flagImportMapping  string
	flagExcludeSchemas string
	flagConfigFile     string
	flagAliasTypes     bool
	flagPrintVersion   bool
)

type configuration struct {
	PackageName     string            `yaml:"package"`
	GenerateTargets []string          `yaml:"generate"`
	OutputFile      string            `yaml:"output"`
	IncludeTags     []string          `yaml:"include-tags"`
	ExcludeTags     []string          `yaml:"exclude-tags"`
	TemplatesDir    string            `yaml:"templates"`
	ImportMapping   map[string]string `yaml:"import-mapping"`
	ExcludeSchemas  []string          `yaml:"exclude-schemas"`
}

func main() {

	flag.StringVar(&flagPackageName, "package", "", "The package name for generated code")
	flag.StringVar(&flagGenerate, "generate", "types,client,server,spec",
		`Comma-separated list of code to generate; valid options: "types", "client", "chi-server", "server", "gin", "spec", "skip-fmt", "skip-prune"`)
	flag.StringVar(&flagOutputFile, "o", "", "Where to output generated code, stdout is default")
	flag.StringVar(&flagIncludeTags, "include-tags", "", "Only include operations with the given tags. Comma-separated list of tags.")
	flag.StringVar(&flagExcludeTags, "exclude-tags", "", "Exclude operations that are tagged with the given tags. Comma-separated list of tags.")
	flag.StringVar(&flagTemplatesDir, "templates", "", "Path to directory containing user templates")
	flag.StringVar(&flagImportMapping, "import-mapping", "", "A dict from the external reference to golang package path")
	flag.StringVar(&flagExcludeSchemas, "exclude-schemas", "", "A comma separated list of schemas which must be excluded from generation")
	flag.StringVar(&flagConfigFile, "config", "", "a YAML config file that controls oapi-codegen behavior")
	flag.BoolVar(&flagAliasTypes, "alias-types", false, "Alias type declarations of possible")
	flag.BoolVar(&flagPrintVersion, "version", false, "when specified, print version and exit")
	flag.Parse()

	if flagPrintVersion {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Fprintln(os.Stderr, "error reading build info")
			os.Exit(1)
		}
		fmt.Println(bi.Main.Path + "/cmd/oapi-codegen")
		fmt.Println(bi.Main.Version)
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Please specify a path to a OpenAPI 3.0 spec file")
		os.Exit(1)
	}

	cfg := configFromFlags()

	// If the package name has not been specified, we will use the name of the
	// swagger file.
	if cfg.PackageName == "" {
		path := flag.Arg(0)
		baseName := filepath.Base(path)
		// Split the base name on '.' to get the first part of the file.
		nameParts := strings.Split(baseName, ".")
		cfg.PackageName = codegen.ToCamelCase(nameParts[0])
	}

	opts := codegen.Options{
		AliasTypes: flagAliasTypes,
	}
	for _, g := range cfg.GenerateTargets {
		switch g {
		case "client":
			opts.GenerateClient = true
		case "chi-server":
			opts.GenerateChiServer = true
		case "server":
			opts.GenerateEchoServer = true
		case "gin":
			opts.GenerateGinServer = true
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

	opts.IncludeTags = cfg.IncludeTags
	opts.ExcludeTags = cfg.ExcludeTags
	opts.ExcludeSchemas = cfg.ExcludeSchemas

	if opts.GenerateEchoServer && opts.GenerateChiServer {
		errExit("can not specify both server and chi-server targets simultaneously")
	}

	swagger, err := util.LoadSwagger(flag.Arg(0))
	if err != nil {
		errExit("error loading swagger spec in %s\n: %s", flag.Arg(0), err)
	}

	templates, err := loadTemplateOverrides(cfg.TemplatesDir)
	if err != nil {
		errExit("error loading template overrides: %s\n", err)
	}
	opts.UserTemplates = templates

	opts.ImportMapping = cfg.ImportMapping

	code, err := codegen.Generate(swagger, cfg.PackageName, opts)
	if err != nil {
		errExit("error generating code: %s\n", err)
	}

	if cfg.OutputFile != "" {
		err = ioutil.WriteFile(cfg.OutputFile, []byte(code), 0644)
		if err != nil {
			errExit("error writing generated code to file: %s", err)
		}
	} else {
		fmt.Print(code)
	}
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
		// Recursively load subdirectory files, using the path relative to the templates
		// directory as the key. This allows for overriding the files in the service-specific
		// directories (e.g. echo, chi, etc.).
		if f.IsDir() {
			subFiles, err := loadTemplateOverrides(path.Join(templatesDir, f.Name()))
			if err != nil {
				return nil, err
			}
			for subDir, subFile := range subFiles {
				templates[path.Join(f.Name(), subDir)] = subFile
			}
			continue
		}
		data, err := ioutil.ReadFile(path.Join(templatesDir, f.Name()))
		if err != nil {
			return nil, err
		}
		templates[f.Name()] = string(data)
	}

	return templates, nil
}

// configFromFlags parses the flags and the config file. Anything which is
// a zerovalue in the configuration file will be replaced with the flag
// value, this means that the config file overrides flag values.
func configFromFlags() *configuration {
	var cfg configuration

	// Load the configuration file first.
	if flagConfigFile != "" {
		f, err := os.Open(flagConfigFile)
		if err != nil {
			errExit("failed to open config file with error: %v\n", err)
		}
		defer f.Close()
		err = yaml.NewDecoder(f).Decode(&cfg)
		if err != nil {
			errExit("error parsing config file: %v\n", err)
		}
	}

	if cfg.PackageName == "" {
		cfg.PackageName = flagPackageName
	}
	if cfg.GenerateTargets == nil {
		cfg.GenerateTargets = util.ParseCommandLineList(flagGenerate)
	}
	if cfg.IncludeTags == nil {
		cfg.IncludeTags = util.ParseCommandLineList(flagIncludeTags)
	}
	if cfg.ExcludeTags == nil {
		cfg.ExcludeTags = util.ParseCommandLineList(flagExcludeTags)
	}
	if cfg.TemplatesDir == "" {
		cfg.TemplatesDir = flagTemplatesDir
	}
	if cfg.ImportMapping == nil && flagImportMapping != "" {
		var err error
		cfg.ImportMapping, err = util.ParseCommandlineMap(flagImportMapping)
		if err != nil {
			errExit("error parsing import-mapping: %s\n", err)
		}
	}
	if cfg.ExcludeSchemas == nil {
		cfg.ExcludeSchemas = util.ParseCommandLineList(flagExcludeSchemas)
	}
	if cfg.OutputFile == "" {
		cfg.OutputFile = flagOutputFile
	}
	return &cfg
}
