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
	"os/exec"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

func errExit(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

var (
	flagOutputFile     string
	flagConfigFile     string
	flagOldConfigStyle bool
	flagOutputConfig   bool
	flagPrintVersion   bool
	flagPackageName    string
	flagPrintUsage     bool
	flagGenerate       string
	flagTemplatesDir   string

	// Deprecated: The options below will be removed in a future
	// release. Please use the new config file format.
	flagIncludeTags         string
	flagExcludeTags         string
	flagIncludeOperationIDs string
	flagExcludeOperationIDs string
	flagImportMapping       string
	flagExcludeSchemas      string
	flagResponseTypeSuffix  string
	flagAliasTypes          bool
	flagInitialismOverrides bool
)

type configuration struct {
	codegen.Configuration `yaml:",inline"`

	// OutputFile is the filename to output.
	OutputFile string `yaml:"output,omitempty"`
}

// oldConfiguration is deprecated. Please add no more flags here. It is here
// for backwards compatibility, and it will be removed in the future.
type oldConfiguration struct {
	PackageName         string                       `yaml:"package"`
	GenerateTargets     []string                     `yaml:"generate"`
	OutputFile          string                       `yaml:"output"`
	IncludeTags         []string                     `yaml:"include-tags"`
	ExcludeTags         []string                     `yaml:"exclude-tags"`
	IncludeOperationIDs []string                     `yaml:"include-operation-ids"`
	ExcludeOperationIDs []string                     `yaml:"exclude-operation-ids"`
	TemplatesDir        string                       `yaml:"templates"`
	ImportMapping       map[string]string            `yaml:"import-mapping"`
	ExcludeSchemas      []string                     `yaml:"exclude-schemas"`
	ResponseTypeSuffix  string                       `yaml:"response-type-suffix"`
	Compatibility       codegen.CompatibilityOptions `yaml:"compatibility"`
}

// noVCSVersionOverride allows overriding the version of the application for cases where no Version Control System (VCS) is available when building, for instance when using a Nix derivation.
// See documentation for how to use it in examples/no-vcs-version-override/README.md
var noVCSVersionOverride string

func main() {
	flag.StringVar(&flagOutputFile, "o", "", "Where to output generated code, stdout is default.")
	flag.BoolVar(&flagOldConfigStyle, "old-config-style", false, "Whether to use the older style config file format.")
	flag.BoolVar(&flagOutputConfig, "output-config", false, "When true, outputs a configuration file for oapi-codegen using current settings.")
	flag.StringVar(&flagConfigFile, "config", "", "A YAML config file that controls oapi-codegen behavior.")
	flag.BoolVar(&flagPrintVersion, "version", false, "When specified, print version and exit.")
	flag.StringVar(&flagPackageName, "package", "", "The package name for generated code.")
	flag.BoolVar(&flagPrintUsage, "help", false, "Show this help and exit.")
	flag.BoolVar(&flagPrintUsage, "h", false, "Same as -help.")

	// All flags below are deprecated, and will be removed in a future release. Please do not
	// update their behavior.
	flag.StringVar(&flagGenerate, "generate", "types,client,server,spec",
		`Comma-separated list of code to generate; valid options: "types", "client", "chi-server", "server", "gin", "gorilla", "spec", "skip-fmt", "skip-prune", "fiber", "iris", "std-http".`)
	flag.StringVar(&flagIncludeTags, "include-tags", "", "Only include operations with the given tags. Comma-separated list of tags.")
	flag.StringVar(&flagExcludeTags, "exclude-tags", "", "Exclude operations that are tagged with the given tags. Comma-separated list of tags.")
	flag.StringVar(&flagIncludeOperationIDs, "include-operation-ids", "", "Only include operations with the given operation-ids. Comma-separated list of operation-ids.")
	flag.StringVar(&flagExcludeOperationIDs, "exclude-operation-ids", "", "Exclude operations with the given operation-ids. Comma-separated list of operation-ids.")
	flag.StringVar(&flagTemplatesDir, "templates", "", "Path to directory containing user templates.")
	flag.StringVar(&flagImportMapping, "import-mapping", "", "A dict from the external reference to golang package path.")
	flag.StringVar(&flagExcludeSchemas, "exclude-schemas", "", "A comma separated list of schemas which must be excluded from generation.")
	flag.StringVar(&flagResponseTypeSuffix, "response-type-suffix", "", "The suffix used for responses types.")
	flag.BoolVar(&flagAliasTypes, "alias-types", false, "Alias type declarations if possible.")
	flag.BoolVar(&flagInitialismOverrides, "initialism-overrides", false, "Use initialism overrides.")

	flag.Parse()

	if flagPrintUsage {
		flag.Usage()
		os.Exit(0)
	}

	if flagPrintVersion {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Fprintln(os.Stderr, "error reading build info")
			os.Exit(1)
		}
		fmt.Println(bi.Main.Path + "/cmd/oapi-codegen")
		version := bi.Main.Version
		if len(noVCSVersionOverride) > 0 {
			version = noVCSVersionOverride
		}
		fmt.Println(version)
		return
	}

	if flag.NArg() < 1 {
		errExit("Please specify a path to a OpenAPI 3.0 spec file\n")
	} else if flag.NArg() > 1 {
		errExit("Only one OpenAPI 3.0 spec file is accepted and it must be the last CLI argument\n")
	}

	// We will try to infer whether the user has an old-style config, or a new
	// style. Start with the command line argument. If it's true, we know it's
	// old config style.
	var oldConfigStyle *bool
	if flagOldConfigStyle {
		oldConfigStyle = &flagOldConfigStyle
	}

	// We don't know yet, so keep looking. Try to parse the configuration file,
	// if given.
	if oldConfigStyle == nil && (flagConfigFile != "") {
		configFile, err := os.ReadFile(flagConfigFile)
		if err != nil {
			errExit("error reading config file '%s': %v\n", flagConfigFile, err)
		}
		var oldConfig oldConfiguration
		oldErr := yaml.UnmarshalStrict(configFile, &oldConfig)

		var newConfig configuration
		newErr := yaml.UnmarshalStrict(configFile, &newConfig)

		// If one of the two files parses, but the other fails, we know the
		// answer.
		if oldErr != nil && newErr == nil {
			f := false
			oldConfigStyle = &f
		} else if oldErr == nil && newErr != nil {
			t := true
			oldConfigStyle = &t
		} else if oldErr != nil && newErr != nil {
			errExit("error parsing configuration style as old version or new version\n\nerror when parsing using old config version:\n%v\n\nerror when parsing using new config version:\n%v\n", oldErr, newErr)
		}
		// Else we fall through, and we still don't know, so we need to infer it from flags.
	}

	if oldConfigStyle == nil {
		// If any deprecated flag is present, and config file structure is unknown,
		// the presence of the deprecated flag means we must be using the old
		// config style. It should work correctly if we go down the old path,
		// even if we have a simple config file readable as both types.
		deprecatedFlagNames := map[string]bool{
			"include-tags":         true,
			"exclude-tags":         true,
			"import-mapping":       true,
			"exclude-schemas":      true,
			"response-type-suffix": true,
			"alias-types":          true,
		}
		hasDeprecatedFlag := false
		flag.Visit(func(f *flag.Flag) {
			if deprecatedFlagNames[f.Name] {
				hasDeprecatedFlag = true
			}
		})
		if hasDeprecatedFlag {
			t := true
			oldConfigStyle = &t
		} else {
			f := false
			oldConfigStyle = &f
		}
	}

	var opts configuration
	if !*oldConfigStyle {
		// We simply read the configuration from disk.
		if flagConfigFile != "" {
			buf, err := os.ReadFile(flagConfigFile)
			if err != nil {
				errExit("error reading config file '%s': %v\n", flagConfigFile, err)
			}
			err = yaml.Unmarshal(buf, &opts)
			if err != nil {
				errExit("error parsing'%s' as YAML: %v\n", flagConfigFile, err)
			}
		} else {
			// In the case where no config file is provided, we assume some
			// defaults, so that when this is invoked very simply, it's similar
			// to old behavior.
			opts = configuration{
				Configuration: codegen.Configuration{
					Generate: codegen.GenerateOptions{
						EchoServer:   true,
						Client:       true,
						Models:       true,
						EmbeddedSpec: true,
					},
				},
				OutputFile: flagOutputFile,
			}
		}

		if err := updateConfigFromFlags(&opts); err != nil {
			errExit("error processing flags: %v\n", err)
		}
	} else {
		var oldConfig oldConfiguration
		if flagConfigFile != "" {
			buf, err := os.ReadFile(flagConfigFile)
			if err != nil {
				errExit("error reading config file '%s': %v\n", flagConfigFile, err)
			}
			err = yaml.Unmarshal(buf, &oldConfig)
			if err != nil {
				errExit("error parsing'%s' as YAML: %v\n", flagConfigFile, err)
			}
		}
		var err error
		opts, err = newConfigFromOldConfig(oldConfig)
		if err != nil {
			flag.PrintDefaults()
			errExit("error creating new config from old config: %v\n", err)
		}

	}

	// Ensure default values are set if user hasn't specified some needed
	// fields.
	opts.Configuration = opts.UpdateDefaults()

	if err := detectPackageName(&opts); err != nil {
		errExit("%s\n", err)
	}

	// Now, ensure that the config options are valid.
	if err := opts.Validate(); err != nil {
		errExit("configuration error: %v\n", err)
	}

	if warnings := opts.Generate.Warnings(); len(warnings) > 0 {
		out := "WARNING: A number of warning(s) were returned when validating the GenerateOptions:"
		for k, v := range warnings {
			out += "\n- " + k + ": " + v
		}

		_, _ = fmt.Fprint(os.Stderr, out)
	}

	// If the user asked to output configuration, output it to stdout and exit
	if flagOutputConfig {
		buf, err := yaml.Marshal(opts)
		if err != nil {
			errExit("error YAML marshaling configuration: %v\n", err)
		}
		fmt.Print(string(buf))
		return
	}

	overlayOpts := util.LoadSwaggerWithOverlayOpts{
		Path: opts.OutputOptions.Overlay.Path,
		// default to strict, but can be overridden
		Strict: true,
	}

	if opts.OutputOptions.Overlay.Strict != nil {
		overlayOpts.Strict = *opts.OutputOptions.Overlay.Strict
	}

	swagger, err := util.LoadSwaggerWithOverlay(flag.Arg(0), overlayOpts)
	if err != nil {
		errExit("error loading swagger spec in %s\n: %s\n", flag.Arg(0), err)
	}

	if strings.HasPrefix(swagger.OpenAPI, "3.1.") {
		fmt.Fprintln(os.Stderr, "WARNING: You are using an OpenAPI 3.1.x specification, which is not yet supported by oapi-codegen (https://github.com/oapi-codegen/oapi-codegen/issues/373) and so some functionality may not be available. Until oapi-codegen supports OpenAPI 3.1, it is recommended to downgrade your spec to 3.0.x")
	}

	if len(noVCSVersionOverride) > 0 {
		opts.NoVCSVersionOverride = &noVCSVersionOverride
	}

	code, err := codegen.Generate(swagger, opts.Configuration)
	if err != nil {
		errExit("error generating code: %s\n", err)
	}

	if opts.OutputFile != "" {
		if err := os.MkdirAll(filepath.Dir(opts.OutputFile), 0o755); err != nil {
			errExit("error unable to create directory: %s\n", err)
		}
		err = os.WriteFile(opts.OutputFile, []byte(code), 0o644)
		if err != nil {
			errExit("error writing generated code to file: %s\n", err)
		}
	} else {
		fmt.Print(code)
	}
}

func loadTemplateOverrides(templatesDir string) (map[string]string, error) {
	templates := make(map[string]string)

	if templatesDir == "" {
		return templates, nil
	}

	files, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		// Recursively load subdirectory files, using the path relative to the templates
		// directory as the key. This allows for overriding the files in the service-specific
		// directories (e.g. echo, chi, fiber, etc.).
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
		data, err := os.ReadFile(path.Join(templatesDir, f.Name()))
		if err != nil {
			return nil, err
		}
		templates[f.Name()] = string(data)
	}

	return templates, nil
}

// detectPackageName detects and sets PackageName if not already set.
func detectPackageName(cfg *configuration) error {
	if cfg.PackageName != "" {
		return nil
	}

	if cfg.OutputFile != "" {
		// Determine from the package name of the output file.
		dir := filepath.Dir(cfg.PackageName)
		cmd := exec.Command("go", "list", "-f", "{{.Name}}", dir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			outStr := string(out)
			switch {
			case strings.Contains(outStr, "expected 'package', found 'EOF'"):
				// Redirecting the output to current directory which hasn't
				// written anything yet, ignore.
			case strings.HasPrefix(outStr, "no Go files in"):
				// No go files yet, ignore.
			default:
				// Unexpected failure report.
				return fmt.Errorf("detect package name for %q output: %q: %w", dir, string(out), err)
			}
		} else {
			cfg.PackageName = string(out)
			return nil
		}
	}

	// Fallback to determining from the spec file name.
	parts := strings.Split(filepath.Base(flag.Arg(0)), ".")
	cfg.PackageName = codegen.LowercaseFirstCharacter(codegen.ToCamelCase(parts[0]))

	return nil
}

// updateConfigFromFlags updates a loaded configuration from flags. Flags
// override anything in the file. We generate errors for any unsupported
// command line flags.
func updateConfigFromFlags(cfg *configuration) error {
	if flagPackageName != "" {
		cfg.PackageName = flagPackageName
	}

	if flagGenerate != "types,client,server,spec" {
		// Override generation and output options from generate command line flag.
		if err := generationTargets(&cfg.Configuration, util.ParseCommandLineList(flagGenerate)); err != nil {
			return err
		}
	}
	if flagIncludeTags != "" {
		cfg.OutputOptions.IncludeTags = util.ParseCommandLineList(flagIncludeTags)
	}
	if flagExcludeTags != "" {
		cfg.OutputOptions.ExcludeTags = util.ParseCommandLineList(flagExcludeTags)
	}
	if flagIncludeOperationIDs != "" {
		cfg.OutputOptions.IncludeOperationIDs = util.ParseCommandLineList(flagIncludeOperationIDs)
	}
	if flagExcludeOperationIDs != "" {
		cfg.OutputOptions.ExcludeOperationIDs = util.ParseCommandLineList(flagExcludeOperationIDs)
	}

	if flagTemplatesDir != "" {
		templates, err := loadTemplateOverrides(flagTemplatesDir)
		if err != nil {
			return fmt.Errorf("load templates from %q: %w", flagTemplatesDir, err)
		}
		cfg.OutputOptions.UserTemplates = templates
	}
	if flagImportMapping != "" {
		var err error
		cfg.ImportMapping, err = util.ParseCommandlineMap(flagImportMapping)
		if err != nil {
			return err
		}
	}
	if flagExcludeSchemas != "" {
		cfg.OutputOptions.ExcludeSchemas = util.ParseCommandLineList(flagExcludeSchemas)
	}
	if flagResponseTypeSuffix != "" {
		cfg.OutputOptions.ResponseTypeSuffix = flagResponseTypeSuffix
	}
	if flagAliasTypes {
		return fmt.Errorf("--alias-types isn't supported any more")
	}

	if cfg.OutputFile == "" {
		cfg.OutputFile = flagOutputFile
	}

	cfg.OutputOptions.InitialismOverrides = flagInitialismOverrides

	return nil
}

// updateOldConfigFromFlags parses the flags and the config file. Anything which is
// a zerovalue in the configuration file will be replaced with the flag
// value, this means that the config file overrides flag values.
func updateOldConfigFromFlags(cfg oldConfiguration) oldConfiguration {
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
	return cfg
}

// generationTargets sets cfg options based on the generation targets.
func generationTargets(cfg *codegen.Configuration, targets []string) error {
	opts := codegen.GenerateOptions{} // Blank to start with.
	for _, opt := range targets {
		switch opt {
		case "iris", "iris-server":
			opts.IrisServer = true
		case "chi-server", "chi":
			opts.ChiServer = true
		case "fiber-server", "fiber":
			opts.FiberServer = true
		case "server", "echo-server", "echo":
			opts.EchoServer = true
		case "gin", "gin-server":
			opts.GinServer = true
		case "gorilla", "gorilla-server":
			opts.GorillaServer = true
		case "std-http", "std-http-server":
			opts.StdHTTPServer = true
		case "strict-server":
			opts.Strict = true
		case "client":
			opts.Client = true
		case "types", "models":
			opts.Models = true
		case "spec", "embedded-spec":
			opts.EmbeddedSpec = true
		case "skip-fmt":
			cfg.OutputOptions.SkipFmt = true
		case "skip-prune":
			cfg.OutputOptions.SkipPrune = true
		default:
			return fmt.Errorf("unknown generate option %q", opt)
		}
	}
	cfg.Generate = opts

	return nil
}

func newConfigFromOldConfig(c oldConfiguration) (configuration, error) {
	// Take flags into account.
	cfg := updateOldConfigFromFlags(c)

	// Now, copy over field by field, translating flags and old values as
	// necessary.
	opts := codegen.Configuration{
		PackageName: cfg.PackageName,
	}
	opts.OutputOptions.ResponseTypeSuffix = flagResponseTypeSuffix

	if err := generationTargets(&opts, cfg.GenerateTargets); err != nil {
		return configuration{}, fmt.Errorf("generation targets: %w", err)
	}

	opts.OutputOptions.IncludeTags = cfg.IncludeTags
	opts.OutputOptions.ExcludeTags = cfg.ExcludeTags
	opts.OutputOptions.ExcludeSchemas = cfg.ExcludeSchemas

	templates, err := loadTemplateOverrides(cfg.TemplatesDir)
	if err != nil {
		return configuration{}, fmt.Errorf("loading template overrides: %w", err)
	}
	opts.OutputOptions.UserTemplates = templates

	opts.ImportMapping = cfg.ImportMapping

	opts.Compatibility = cfg.Compatibility

	return configuration{
		Configuration: opts,
		OutputFile:    cfg.OutputFile,
	}, nil
}
