package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"

	"gopkg.in/yaml.v3"

	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen"
)

func main() {
	configPath := flag.String("config", "", "path to configuration file")
	flagPackage := flag.String("package", "", "Go package name for generated code")
	flagOutput := flag.String("output", "", "output file path (default: <spec-basename>.gen.go)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <spec-path>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  spec-path    path to OpenAPI spec file\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	specPath := flag.Arg(0)

	// Parse the OpenAPI spec
	specData, err := os.ReadFile(specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading spec: %v\n", err)
		os.Exit(1)
	}

	// Configure libopenapi to skip resolving external references.
	// We handle external $refs via import mappings — the referenced specs
	// don't need to be fetched or parsed. See pb33f/libopenapi#519.
	docConfig := datamodel.NewDocumentConfiguration()
	docConfig.SkipExternalRefResolution = true

	doc, err := libopenapi.NewDocumentWithConfiguration(specData, docConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing spec: %v\n", err)
		os.Exit(1)
	}

	// Parse config if provided, otherwise use empty config
	var cfg codegen.Configuration
	if *configPath != "" {
		configData, err := os.ReadFile(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading config: %v\n", err)
			os.Exit(1)
		}
		if err := yaml.Unmarshal(configData, &cfg); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing config: %v\n", err)
			os.Exit(1)
		}
	}

	// Flags override config file values
	if *flagPackage != "" {
		cfg.PackageName = *flagPackage
	}
	if *flagOutput != "" {
		cfg.Output = *flagOutput
	}

	// Default output to <spec-basename>.gen.go
	if cfg.Output == "" {
		base := filepath.Base(specPath)
		ext := filepath.Ext(base)
		cfg.Output = strings.TrimSuffix(base, ext) + ".gen.go"
	}

	// Default package name from output file
	if cfg.PackageName == "" {
		cfg.PackageName = "api"
	}

	// Generate code
	code, err := codegen.Generate(doc, specData, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating code: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(cfg.Output, []byte(code), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s\n", cfg.Output)
}
