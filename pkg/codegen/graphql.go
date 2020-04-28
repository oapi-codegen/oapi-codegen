package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/deepmap/oapi-codegen/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/integralist/go-findroot/find"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/v2/ast"
)

var rootPath string

func init() {
	rep, err := find.Repo()
	if err != nil {
		panic(err)
	}
	rootPath = rep.Path
}

func GenerateGraphQL(t *template.Template, swagger *openapi3.Swagger) (string, error) {
	// if err != nil {
	// 	return "", err
	// }
	ops, err := OperationDefinitions(swagger)
	if err != nil {
		return "", err
	}
	s, err := GenerateGraphQLSchema(t, ops)
	if err != nil {
		return "", errors.Wrap(err, "Error generating graphql schema")
	}
	s, err = GenerateGraphQLTypes(t, swagger, ops)
	if err != nil {
		return "", errors.Wrap(err, "Error generating graphQL types")
	}
	s, err = GenerateGraphQLInputs(t, swagger, ops)
	if err != nil {
		return "", errors.Wrap(err, "Error generating graphQL inputs")
	}
	s, err = GenerateScalars(t)
	if err != nil {
		return s, err
	}
	s, err = GenerateResolvers(t, ops)
	if err != nil {
		return s, err
	}
	err = GenerateGQLGen(swagger, ops)
	if err != nil {
		return "", errors.Wrap(err, "Error generating gqlgen")
	}
	return s, nil
}

func GenerateGraphQLSchema(t *template.Template, ops []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	type Config struct {
		HasQuery    bool
		HasMutation bool
		Operations  []OperationDefinition
	}
	var hasQuery = false
	var hasMutation = false
	for _, o := range ops {
		if o.Method == "GET" {
			hasQuery = true
		} else {
			hasMutation = true
		}
	}
	cfg := Config{
		HasQuery:    hasQuery,
		HasMutation: hasMutation,
		Operations:  ops,
	}
	err := t.ExecuteTemplate(w, "graphql-root.tmpl", cfg)
	if err != nil {
		return "", errors.Wrap(err, "error generating types")
	}
	err = w.Flush()
	if err != nil {
		return "", errors.Wrap(err, "error flushing output buffer for types")
	}
	err = ioutil.WriteFile(rootPath+"/graph/schema.gql", buf.Bytes(), 0644)
	if err != nil {
		return "", err
	}
	s := buf.String()
	println(s)
	return buf.String(), nil
}

type TypeConfig struct {
	Kind  string
	Types map[string]TypeDefinition
}

func GetTypeDefinitions(swagger *openapi3.Swagger, ops []OperationDefinition) (map[string]TypeDefinition, error) {
	// Aggregate all type definitions needed
	schemaTypes, err := GenerateTypesForSchemas(swagger.Components.Schemas)
	if err != nil {
		return nil, errors.Wrap(err, "error generating Go types for component schemas")
	}

	paramTypes, err := GenerateTypesForParameters(swagger.Components.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "error generating Go types for component parameters")
	}
	allTypes := append(schemaTypes, paramTypes...)

	responseTypes, err := GenerateTypesForResponses(swagger.Components.Responses)
	if err != nil {
		return nil, errors.Wrap(err, "error generating Go types for component responses")
	}
	allTypes = append(allTypes, responseTypes...)

	bodyTypes, err := GenerateTypesForRequestBodies(swagger.Components.RequestBodies)
	if err != nil {
		return nil, errors.Wrap(err, "error generating Go types for component request bodies")
	}
	allTypes = append(allTypes, bodyTypes...)

	customResponses, err := GenerateGraphQLCustomResponses(ops)
	if err != nil {
		return nil, errors.Wrap(err, "error generating graphql custome responses")
	}
	allTypes = append(allTypes, customResponses...)

	atm := map[string]TypeDefinition{}
	for _, td := range allTypes {
		atm[td.TypeName] = td
	}
	return atm, nil
}

func GenerateGraphQLTypes(t *template.Template, swagger *openapi3.Swagger, ops []OperationDefinition) (string, error) {
	allTypes, err := GetTypeDefinitions(swagger, ops)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	config := TypeConfig{
		Kind:  "type",
		Types: allTypes,
	}
	err = t.ExecuteTemplate(w, "graphql-types.tmpl", config)
	if err != nil {
		return "", errors.Wrap(err, "error generating types")
	}
	err = w.Flush()
	if err != nil {
		return "", errors.Wrap(err, "error flushing output buffer for types")
	}
	err = ioutil.WriteFile(rootPath+"/graph/types.gql", buf.Bytes(), 0644)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func CollectGraphQLInputs(ops []OperationDefinition) []TypeDefinition {
	inputTypes := []TypeDefinition{}
	for _, o := range ops {
		for _, b := range o.Bodies {
			inputTypes = append(inputTypes, TypeDefinition{
				TypeName: o.OperationId + b.NameTag + "Body",
				Schema:   b.Schema,
			})
		}
	}
	return inputTypes
}

func recursiveInputs(td TypeDefinition, allTypes map[string]TypeDefinition, tds []TypeDefinition) []TypeDefinition {
	// if there is no properties we have areference to another type
	if td.Schema.Properties == nil {
		td.Schema = allTypes[td.Schema.GoType].Schema
	}
	for _, p := range td.Schema.Properties {
		// remove
		switch strings.TrimPrefix(p.Schema.GoType, "[]") {
		case "", "string", "float", "float32", "float64", "interface{}", "int", "int32", "int64", "bool", "time.Time", "openapi_types.Date":
			continue
		default:
			t := allTypes[p.Schema.GoType]
			tds = append(tds, t)
			tds = recursiveInputs(t, allTypes, tds)
		}
	}
	return tds
}

// GetGraphQLInputs return all typedefinitions needed for graphql inputs
func GetGraphQLInputs(swagger *openapi3.Swagger, ops []OperationDefinition) (map[string]TypeDefinition, error) {
	inputs := map[string]TypeDefinition{}
	allTypes, err := GetTypeDefinitions(swagger, ops)
	if err != nil {
		return inputs, err
	}
	inputTypes := CollectGraphQLInputs(ops)
	for _, i := range inputTypes {
		// get all sub input types for that type
		subInputs := []TypeDefinition{}
		subInputs = recursiveInputs(i, allTypes, subInputs)
		if i.Schema.Properties == nil {
			i.Schema = allTypes[i.Schema.GoType].Schema
		}
		inputs[i.TypeName] = i
		for _, td := range subInputs {
			inputs[td.TypeName] = td
		}
	}
	return inputs, nil
}

func GenerateGraphQLInputs(t *template.Template, swagger *openapi3.Swagger, ops []OperationDefinition) (string, error) {
	inputs, err := GetGraphQLInputs(swagger, ops)
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	config := TypeConfig{
		Kind:  "input",
		Types: inputs,
	}
	err = t.ExecuteTemplate(w, "graphql-types.tmpl", config)
	if err != nil {
		return "", errors.Wrap(err, "error generating types")
	}
	err = w.Flush()
	if err != nil {
		return "", errors.Wrap(err, "error flushing output buffer for types")
	}
	err = ioutil.WriteFile(rootPath+"/graph/inputs.gql", buf.Bytes(), 0644)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateScalars add the needed scalars for graphQL
func GenerateScalars(t *template.Template) (string, error) {
	var s string
	{
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)

		err := t.ExecuteTemplate(w, "graphql-scalar.tmpl", nil)
		if err != nil {
			return "", errors.Wrap(err, "error generating scalar graphql types")
		}
		err = w.Flush()
		if err != nil {
			return "", errors.Wrap(err, "error flushing output buffer for types")
		}
		ioutil.WriteFile(rootPath+"/graph/scalars.gql", buf.Bytes(), 0644)
		s += buf.String()
	}
	{
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)

		err := t.ExecuteTemplate(w, "go-scalars.tmpl", nil)
		if err != nil {
			return "", errors.Wrap(err, "error generating scalar graphql types")
		}
		err = w.Flush()
		if err != nil {
			return "", errors.Wrap(err, "error flushing output buffer for types")
		}
		ioutil.WriteFile(rootPath+"/graph/scalars.go", buf.Bytes(), 0644)
		s += buf.String()
	}
	return s, nil
}

type ResolverConfig struct {
	HasMutation   bool
	HasQuery      bool
	ClientPackage string
	TypePackage   string
	CurrentModule string
	Operations    []OperationDefinition
}

// GenerateResolvers generate resolvers for graphql
func GenerateResolvers(t *template.Template, ops []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	m, err := util.GetCurrentModule()
	if err != nil {
		return "", err
	}
	var hasQuery = false
	var hasMutation = false
	for _, o := range ops {
		if o.Method == "GET" {
			hasQuery = true
		} else {
			hasMutation = true
		}
	}
	cfg := ResolverConfig{
		HasQuery:      hasQuery,
		HasMutation:   hasMutation,
		ClientPackage: "codegen",
		TypePackage:   "codegen",
		CurrentModule: m,
		Operations:    ops,
	}

	err = t.ExecuteTemplate(w, "graphql-resolver.tmpl", cfg)
	if err != nil {
		return "", errors.Wrap(err, "error generating resolver graphql types")
	}
	err = w.Flush()
	if err != nil {
		return "", errors.Wrap(err, "error flushing output buffer for resolvers")
	}
	ioutil.WriteFile(rootPath+"/graph/resolvers.go", buf.Bytes(), 0644)
	return buf.String(), nil
}

// GenerateGraphQLCustomResponses return the response needed to define a graphql schema
func GenerateGraphQLCustomResponses(ops []OperationDefinition) ([]TypeDefinition, error) {
	tds := []TypeDefinition{}
	for _, o := range ops {
		// The valid response is of code 200, @WARNING
		response := o.Spec.Responses["200"]
		if response.Value != nil {
			sortedContentKeys := SortedContentKeys(response.Value.Content)
			for _, contentTypeName := range sortedContentKeys {
				contentType := response.Value.Content[contentTypeName]
				if contentType.Schema != nil && contentType.Schema.Ref == "" && contentType.Schema.Value.Type == "object" {
					responseSchema, err := GenerateGoSchema(contentType.Schema, []string{o.OperationId + "GQLResponse"})
					if err != nil {
						return nil, errors.Wrap(err, fmt.Sprintf("Unable to determine Go type for %s.%s", o.OperationId, contentTypeName))
					}
					td := TypeDefinition{
						TypeName: o.OperationId + "GQLResponse",
						Schema:   responseSchema,
					}
					tds = append(tds, td)
				}
				if contentType.Schema != nil && contentType.Schema.Value.Type == "array" && contentType.Schema.Value.Items.Value.Type == "object" && contentType.Schema.Value.Items.Ref == "" {
					responseSchema, err := GenerateGoSchema(contentType.Schema, []string{o.OperationId + "GQLResponse"})
					if err != nil {
						return nil, errors.Wrap(err, fmt.Sprintf("Unable to determine Go type for %s.%s", o.OperationId, contentTypeName))
					}
					td := TypeDefinition{
						TypeName: o.OperationId + "GQLResponse",
						Schema:   responseSchema,
					}
					tds = append(tds, td)
				}
			}
		}
	}
	return tds, nil
}

// GenerateGQLGen generate the graphql server using github.com/99designs/gqlgen
func GenerateGQLGen(swagger *openapi3.Swagger, ops []OperationDefinition) error {
	m, err := util.GetCurrentModule()
	if err != nil {
		return err
	}
	cfg := config.Config{
		Exec: config.PackageConfig{
			Filename: rootPath + "/graph/main.go",
			Package:  "graph",
		},
		Directives: map[string]config.DirectiveConfig{},
		AutoBind:   []string{m + "/codegen"},
		// Federation: config.PackageConfig{
		// 	Filename: rootPath + "/graph/federation.go",
		// 	Package:  "graph",
		// },
	}
	models := config.TypeMap{
		"Float32": {Model: config.StringList{m + "/graph.Float32"}},
		"Date":    {Model: config.StringList{m + "/graph.Date"}},
	}
	inputTypes, err := GetGraphQLInputs(swagger, ops)
	if err != nil {
		return err
	}
	for _, t := range inputTypes {
		models[t.TypeName+"Input"] = config.TypeMapEntry{Model: config.StringList{m + "/codegen." + t.TypeName}}
	}
	cfg.Models = models
	// @todo, add collected input gql types
	files := []string{"schema.gql", "types.gql", "inputs.gql", "scalars.gql"}
	for _, filename := range files {
		path := rootPath + "/graph/" + filename
		schemaRaw, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "unable to open schema")
		}
		cfg.Sources = append(cfg.Sources, &ast.Source{Name: filename, Input: string(schemaRaw)})
	}
	err = api.Generate(&cfg)
	if err != nil {
		return err
	}
	return nil
}
