package codegen

import (
	"fmt"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

type RefWrapper struct {
	Ref       string
	HasValue  bool
	SourceRef any
}

func walkSwagger(swagger *openapi3.T, doFn func(RefWrapper) (bool, error)) error {
	if swagger == nil {
		return nil
	}

	if swagger.Paths != nil {
		for _, pathItem := range swagger.Paths.Map() {
			_ = walkPathItem(pathItem, doFn)
		}
	}
	for _, webhook := range swagger.Webhooks {
		_ = walkPathItem(webhook, doFn)
	}

	_ = walkComponents(swagger.Components, doFn)

	return nil
}

func walkPathItem(pathItem *openapi3.PathItem, doFn func(RefWrapper) (bool, error)) error {
	if pathItem == nil {
		return nil
	}

	for _, parameter := range pathItem.Parameters {
		_ = walkParameterRef(parameter, doFn)
	}
	for _, operation := range pathItem.Operations() {
		_ = walkOperation(operation, doFn)
	}

	return nil
}

func walkOperation(op *openapi3.Operation, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if op == nil {
		return nil
	}

	for _, param := range op.Parameters {
		_ = walkParameterRef(param, doFn)
	}

	_ = walkRequestBodyRef(op.RequestBody, doFn)

	if op.Responses != nil {
		for _, response := range op.Responses.Map() {
			_ = walkResponseRef(response, doFn)
		}
	}

	for _, callback := range op.Callbacks {
		_ = walkCallbackRef(callback, doFn)
	}

	return nil
}

func walkComponents(components *openapi3.Components, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if components == nil {
		return nil
	}

	for _, schema := range components.Schemas {
		_ = walkSchemaRef(schema, doFn)
	}

	for _, param := range components.Parameters {
		_ = walkParameterRef(param, doFn)
	}

	for _, header := range components.Headers {
		_ = walkHeaderRef(header, doFn)
	}

	for _, requestBody := range components.RequestBodies {
		_ = walkRequestBodyRef(requestBody, doFn)
	}

	for _, response := range components.Responses {
		_ = walkResponseRef(response, doFn)
	}

	for _, securityScheme := range components.SecuritySchemes {
		_ = walkSecuritySchemeRef(securityScheme, doFn)
	}

	for _, example := range components.Examples {
		_ = walkExampleRef(example, doFn)
	}

	for _, link := range components.Links {
		_ = walkLinkRef(link, doFn)
	}

	for _, callback := range components.Callbacks {
		_ = walkCallbackRef(callback, doFn)
	}

	return nil
}

func walkSchemaRef(ref *openapi3.SchemaRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	for _, child := range schemaChildRefs(ref.Value) {
		_ = walkSchemaRef(child, doFn)
	}

	return nil
}

// schemaChildRefs returns every SchemaRef-bearing keyword supported by the
// current kin-openapi Schema type. Keep this list in sync when that type grows.
func schemaChildRefs(schema *openapi3.Schema) []*openapi3.SchemaRef {
	if schema == nil {
		return nil
	}

	refs := make([]*openapi3.SchemaRef, 0,
		len(schema.OneOf)+len(schema.AnyOf)+len(schema.AllOf)+
			len(schema.Properties)+len(schema.PrefixItems)+
			len(schema.PatternProperties)+len(schema.DependentSchemas)+len(schema.Defs)+12)
	refs = append(refs, schema.OneOf...)
	refs = append(refs, schema.AnyOf...)
	refs = append(refs, schema.AllOf...)
	refs = append(refs, schema.Not, schema.Items)
	for _, child := range schema.Properties {
		refs = append(refs, child)
	}
	refs = append(refs, schema.AdditionalProperties.Schema)
	refs = append(refs, schema.PrefixItems...)
	refs = append(refs, schema.Contains)
	for _, child := range schema.PatternProperties {
		refs = append(refs, child)
	}
	for _, child := range schema.DependentSchemas {
		refs = append(refs, child)
	}
	refs = append(refs,
		schema.PropertyNames,
		schema.UnevaluatedItems.Schema,
		schema.UnevaluatedProperties.Schema,
		schema.If,
		schema.Then,
		schema.Else,
	)
	for _, child := range schema.Defs {
		refs = append(refs, child)
	}
	refs = append(refs, schema.ContentSchema)

	return refs
}

func walkParameterRef(ref *openapi3.ParameterRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	_ = walkSchemaRef(ref.Value.Schema, doFn)

	for _, example := range ref.Value.Examples {
		_ = walkExampleRef(example, doFn)
	}

	_ = walkContent(ref.Value.Content, doFn)

	return nil
}

func walkRequestBodyRef(ref *openapi3.RequestBodyRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	_ = walkContent(ref.Value.Content, doFn)

	return nil
}

func walkResponseRef(ref *openapi3.ResponseRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	for _, header := range ref.Value.Headers {
		_ = walkHeaderRef(header, doFn)
	}

	_ = walkContent(ref.Value.Content, doFn)

	for _, link := range ref.Value.Links {
		_ = walkLinkRef(link, doFn)
	}

	return nil
}

func walkContent(content openapi3.Content, doFn func(RefWrapper) (bool, error)) error {
	for _, mediaType := range content {
		if mediaType == nil {
			continue
		}
		_ = walkSchemaRef(mediaType.Schema, doFn)
		_ = walkSchemaRef(mediaType.ItemSchema, doFn)

		for _, example := range mediaType.Examples {
			_ = walkExampleRef(example, doFn)
		}
		for _, encoding := range mediaType.Encoding {
			if encoding == nil {
				continue
			}
			for _, header := range encoding.Headers {
				_ = walkHeaderRef(header, doFn)
			}
		}
	}

	return nil
}

func walkCallbackRef(ref *openapi3.CallbackRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	for _, pathItem := range ref.Value.Map() {
		_ = walkPathItem(pathItem, doFn)
	}

	return nil
}

func walkHeaderRef(ref *openapi3.HeaderRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	_ = walkSchemaRef(ref.Value.Schema, doFn)
	_ = walkContent(ref.Value.Content, doFn)

	return nil
}

func walkSecuritySchemeRef(ref *openapi3.SecuritySchemeRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	// NOTE: `SecuritySchemeRef`s don't contain any children that can contain refs

	return nil
}

func walkLinkRef(ref *openapi3.LinkRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	return nil
}

func walkExampleRef(ref *openapi3.ExampleRef, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.Ref, HasValue: ref.Value != nil, SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}
	if ref.Value == nil {
		return nil
	}

	// NOTE: `ExampleRef`s don't contain any children that can contain refs

	return nil
}

// schemaKeywords are the keywords whose value is, or holds, a schema: the ones
// schemaChildRefs reads. Keep the two in sync.
var schemaKeywords = []string{
	"oneOf", "anyOf", "allOf", "not", "items", "properties", "additionalProperties",
	"prefixItems", "contains", "patternProperties", "dependentSchemas", "propertyNames",
	"unevaluatedItems", "unevaluatedProperties", "if", "then", "else", "$defs", "contentSchema",
}

// refSiblingSchema returns the schema made of the keywords written next to a
// schema $ref, or nil when there are none which could hold a reference.
//
// Since OpenAPI 3.1 a $ref may have sibling keywords, which apply alongside the
// schema it refers to. kin-openapi does not expose them. What it does is make
// the reference's Value a copy of the referenced schema, with one more allOf
// entry holding the siblings, so that entry is what this returns. The only
// trace of the siblings it leaves in the open is that it writes them back when
// marshaling the reference, which is how this tells whether there are any.
func refSiblingSchema(ref *openapi3.SchemaRef, isOpenAPI31OrLater bool) *openapi3.SchemaRef {
	// Before 3.1 a $ref replaces the object it is in, siblings and all.
	if !isOpenAPI31OrLater || ref.Ref == "" || ref.Value == nil || len(ref.Value.AllOf) == 0 {
		return nil
	}

	marshaled, err := ref.MarshalYAML()
	if err != nil {
		return nil
	}
	written, ok := marshaled.(map[string]any)
	if !ok {
		return nil
	}
	if !slices.ContainsFunc(schemaKeywords, func(keyword string) bool {
		_, ok := written[keyword]
		return ok
	}) {
		return nil
	}

	siblings := ref.Value.AllOf[len(ref.Value.AllOf)-1]
	if siblings == nil || siblings.Ref != "" {
		return nil
	}
	return siblings
}

func findComponentRefs(swagger *openapi3.T) []string {
	refs := []string{}
	if swagger == nil {
		return refs
	}
	isOpenAPI31OrLater := swagger.IsOpenAPI31OrLater()

	var record func(ref RefWrapper) (bool, error)
	record = func(ref RefWrapper) (bool, error) {
		if ref.Ref == "" {
			return true, nil
		}
		refs = append(refs, ref.Ref)

		// Stop here: what the reference resolves to is not written in this
		// place. If it is a part of this document, the walk gets to it on its
		// own. If it is a part of another one, the references within it are
		// relative to that document, and would be mistaken for references to
		// the components of this one.
		//
		// The keywords written next to the reference are a part of this
		// document though, wherever the reference itself leads.
		if schemaRef, ok := ref.SourceRef.(*openapi3.SchemaRef); ok {
			_ = walkSchemaRef(refSiblingSchema(schemaRef, isOpenAPI31OrLater), record)
		}
		return false, nil
	}

	_ = walkSwagger(swagger, record)

	return refs
}

func removeOrphanedComponents(swagger *openapi3.T, refs []string) int {
	if swagger.Components == nil {
		return 0
	}

	countRemoved := 0

	for key := range swagger.Components.Schemas {
		ref := fmt.Sprintf("#/components/schemas/%s", key)
		if !slices.Contains(refs, ref) {
			countRemoved++
			delete(swagger.Components.Schemas, key)
		}
	}

	for key := range swagger.Components.Parameters {
		ref := fmt.Sprintf("#/components/parameters/%s", key)
		if !slices.Contains(refs, ref) {
			countRemoved++
			delete(swagger.Components.Parameters, key)
		}
	}

	// securitySchemes are an exception. definitions in securitySchemes
	// are referenced directly by name. and not by $ref

	// for key, _ := range swagger.Components.SecuritySchemes {
	// 	ref := fmt.Sprintf("#/components/securitySchemes/%s", key)
	// 	if !slices.Contains(refs, ref) {
	// 		countRemoved++
	// 		delete(swagger.Components.SecuritySchemes, key)
	// 	}
	// }

	for key := range swagger.Components.RequestBodies {
		ref := fmt.Sprintf("#/components/requestBodies/%s", key)
		if !slices.Contains(refs, ref) {
			countRemoved++
			delete(swagger.Components.RequestBodies, key)
		}
	}

	for key := range swagger.Components.Responses {
		ref := fmt.Sprintf("#/components/responses/%s", key)
		if !slices.Contains(refs, ref) {
			countRemoved++
			delete(swagger.Components.Responses, key)
		}
	}

	for key := range swagger.Components.Headers {
		ref := fmt.Sprintf("#/components/headers/%s", key)
		if !slices.Contains(refs, ref) {
			countRemoved++
			delete(swagger.Components.Headers, key)
		}
	}

	for key := range swagger.Components.Examples {
		ref := fmt.Sprintf("#/components/examples/%s", key)
		if !slices.Contains(refs, ref) {
			countRemoved++
			delete(swagger.Components.Examples, key)
		}
	}

	for key := range swagger.Components.Links {
		ref := fmt.Sprintf("#/components/links/%s", key)
		if !slices.Contains(refs, ref) {
			countRemoved++
			delete(swagger.Components.Links, key)
		}
	}

	for key := range swagger.Components.Callbacks {
		ref := fmt.Sprintf("#/components/callbacks/%s", key)
		if !slices.Contains(refs, ref) {
			countRemoved++
			delete(swagger.Components.Callbacks, key)
		}
	}

	return countRemoved
}

func pruneUnusedComponents(swagger *openapi3.T) {
	for {
		refs := findComponentRefs(swagger)
		countRemoved := removeOrphanedComponents(swagger, refs)
		if countRemoved < 1 {
			break
		}
	}
}
