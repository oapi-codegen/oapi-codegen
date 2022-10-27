package codegen

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type RefWrapper struct {
	Ref       string
	HasValue  bool
	SourceRef interface{}
}

func walkSwagger(swagger *openapi3.T, doFn func(RefWrapper) (bool, error)) error {
	if swagger == nil {
		return nil
	}

	for _, p := range swagger.Paths {
		for _, param := range p.Parameters {
			walkParameterRef(param, doFn)
		}
		for _, op := range p.Operations() {
			walkOperation(op, doFn)
		}
	}

	walkComponents(&swagger.Components, doFn)

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

	for _, response := range op.Responses {
		walkResponseRef(response, doFn)
	}

	for _, callback := range op.Callbacks {
		walkCallbackRef(callback, doFn)
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

	for _, ref := range ref.Value.OneOf {
		walkSchemaRef(ref, doFn)
	}

	for _, ref := range ref.Value.AnyOf {
		walkSchemaRef(ref, doFn)
	}

	for _, ref := range ref.Value.AllOf {
		walkSchemaRef(ref, doFn)
	}

	walkSchemaRef(ref.Value.Not, doFn)
	walkSchemaRef(ref.Value.Items, doFn)

	for _, ref := range ref.Value.Properties {
		walkSchemaRef(ref, doFn)
	}

	walkSchemaRef(ref.Value.AdditionalProperties, doFn)

	return nil
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

	walkSchemaRef(ref.Value.Schema, doFn)

	for _, example := range ref.Value.Examples {
		walkExampleRef(example, doFn)
	}

	for _, mediaType := range ref.Value.Content {
		if mediaType == nil {
			continue
		}
		walkSchemaRef(mediaType.Schema, doFn)

		for _, example := range mediaType.Examples {
			walkExampleRef(example, doFn)
		}
	}

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

	for _, mediaType := range ref.Value.Content {
		if mediaType == nil {
			continue
		}
		walkSchemaRef(mediaType.Schema, doFn)

		for _, example := range mediaType.Examples {
			walkExampleRef(example, doFn)
		}
	}

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
		walkHeaderRef(header, doFn)
	}

	for _, mediaType := range ref.Value.Content {
		if mediaType == nil {
			continue
		}
		walkSchemaRef(mediaType.Schema, doFn)

		for _, example := range mediaType.Examples {
			walkExampleRef(example, doFn)
		}
	}

	for _, link := range ref.Value.Links {
		walkLinkRef(link, doFn)
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

	for _, pathItem := range *ref.Value {
		for _, parameter := range pathItem.Parameters {
			walkParameterRef(parameter, doFn)
		}
		walkOperation(pathItem.Connect, doFn)
		walkOperation(pathItem.Delete, doFn)
		walkOperation(pathItem.Get, doFn)
		walkOperation(pathItem.Head, doFn)
		walkOperation(pathItem.Options, doFn)
		walkOperation(pathItem.Patch, doFn)
		walkOperation(pathItem.Post, doFn)
		walkOperation(pathItem.Put, doFn)
		walkOperation(pathItem.Trace, doFn)
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

	walkSchemaRef(ref.Value.Schema, doFn)

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

func findComponentRefs(swagger *openapi3.T) []string {
	refs := []string{}

	walkSwagger(swagger, func(ref RefWrapper) (bool, error) {
		if ref.Ref != "" {
			refs = append(refs, ref.Ref)
			return false, nil
		}
		return true, nil
	})

	return refs
}

func removeOrphanedComponents(swagger *openapi3.T, refs []string) int {
	countRemoved := 0

	for key := range swagger.Components.Schemas {
		ref := fmt.Sprintf("#/components/schemas/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			delete(swagger.Components.Schemas, key)
		}
	}

	for key := range swagger.Components.Parameters {
		ref := fmt.Sprintf("#/components/parameters/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			delete(swagger.Components.Parameters, key)
		}
	}

	// securitySchemes are an exception. definitions in securitySchemes
	// are referenced directly by name. and not by $ref

	// for key, _ := range swagger.Components.SecuritySchemes {
	// 	ref := fmt.Sprintf("#/components/securitySchemes/%s", key)
	// 	if !stringInSlice(ref, refs) {
	// 		countRemoved++
	// 		delete(swagger.Components.SecuritySchemes, key)
	// 	}
	// }

	for key := range swagger.Components.RequestBodies {
		ref := fmt.Sprintf("#/components/requestBodies/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			delete(swagger.Components.RequestBodies, key)
		}
	}

	for key := range swagger.Components.Responses {
		ref := fmt.Sprintf("#/components/responses/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			delete(swagger.Components.Responses, key)
		}
	}

	for key := range swagger.Components.Headers {
		ref := fmt.Sprintf("#/components/headers/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			delete(swagger.Components.Headers, key)
		}
	}

	for key := range swagger.Components.Examples {
		ref := fmt.Sprintf("#/components/examples/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			delete(swagger.Components.Examples, key)
		}
	}

	for key := range swagger.Components.Links {
		ref := fmt.Sprintf("#/components/links/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			delete(swagger.Components.Links, key)
		}
	}

	for key := range swagger.Components.Callbacks {
		ref := fmt.Sprintf("#/components/callbacks/%s", key)
		if !stringInSlice(ref, refs) {
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
