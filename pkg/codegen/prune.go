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
			if err := walkParameterRef(param, doFn); err != nil {
				return err
			}
		}
		for _, op := range p.Operations() {
			if err := walkOperation(op, doFn); err != nil {
				return err
			}
		}
	}

	return walkComponents(&swagger.Components, doFn)
}

func walkOperation(op *openapi3.Operation, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if op == nil {
		return nil
	}

	for _, param := range op.Parameters {
		if err := walkParameterRef(param, doFn); err != nil {
			return err
		}
	}

	if err := walkRequestBodyRef(op.RequestBody, doFn); err != nil {
		return err
	}

	for _, response := range op.Responses {
		if err := walkResponseRef(response, doFn); err != nil {
			return err
		}
	}

	for _, callback := range op.Callbacks {
		if err := walkCallbackRef(callback, doFn); err != nil {
			return err
		}
	}

	return nil
}

func walkComponents(components *openapi3.Components, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if components == nil {
		return nil
	}

	for _, schema := range components.Schemas {
		if err := walkSchemaRef(schema, doFn); err != nil {
			return err
		}
	}

	for _, param := range components.Parameters {
		if err := walkParameterRef(param, doFn); err != nil {
			return err
		}
	}

	for _, header := range components.Headers {
		if err := walkHeaderRef(header, doFn); err != nil {
			return err
		}
	}

	for _, requestBody := range components.RequestBodies {
		if err := walkRequestBodyRef(requestBody, doFn); err != nil {
			return err
		}
	}

	for _, response := range components.Responses {
		if err := walkResponseRef(response, doFn); err != nil {
			return err
		}
	}

	for _, securityScheme := range components.SecuritySchemes {
		if err := walkSecuritySchemeRef(securityScheme, doFn); err != nil {
			return err
		}
	}

	for _, example := range components.Examples {
		if err := walkExampleRef(example, doFn); err != nil {
			return err
		}
	}

	for _, link := range components.Links {
		if err := walkLinkRef(link, doFn); err != nil {
			return err
		}
	}

	for _, callback := range components.Callbacks {
		if err := walkCallbackRef(callback, doFn); err != nil {
			return err
		}
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
		if err := walkSchemaRef(ref, doFn); err != nil {
			return err
		}
	}

	for _, ref := range ref.Value.AnyOf {
		if err := walkSchemaRef(ref, doFn); err != nil {
			return err
		}
	}

	for _, ref := range ref.Value.AllOf {
		if err := walkSchemaRef(ref, doFn); err != nil {
			return err
		}
	}

	if err := walkSchemaRef(ref.Value.Not, doFn); err != nil {
		return err
	}
	if err := walkSchemaRef(ref.Value.Items, doFn); err != nil {
		return err
	}

	for _, ref := range ref.Value.Properties {
		if err := walkSchemaRef(ref, doFn); err != nil {
			return err
		}
	}

	return walkSchemaRef(ref.Value.AdditionalProperties, doFn)
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

	if err := walkSchemaRef(ref.Value.Schema, doFn); err != nil {
		return err
	}

	for _, example := range ref.Value.Examples {
		if err := walkExampleRef(example, doFn); err != nil {
			return err
		}
	}

	for _, mediaType := range ref.Value.Content {
		if mediaType == nil {
			continue
		}

		if err := walkSchemaRef(mediaType.Schema, doFn); err != nil {
			return err
		}

		for _, example := range mediaType.Examples {
			if err := walkExampleRef(example, doFn); err != nil {
				return err
			}
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
		if err := walkSchemaRef(mediaType.Schema, doFn); err != nil {
			return err
		}

		for _, example := range mediaType.Examples {
			if err := walkExampleRef(example, doFn); err != nil {
				return err
			}
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
		if err := walkHeaderRef(header, doFn); err != nil {
			return err
		}
	}

	for _, mediaType := range ref.Value.Content {
		if mediaType == nil {
			continue
		}

		if err := walkSchemaRef(mediaType.Schema, doFn); err != nil {
			return err
		}

		for _, example := range mediaType.Examples {
			if err := walkExampleRef(example, doFn); err != nil {
				return err
			}
		}
	}

	for _, link := range ref.Value.Links {
		if err := walkLinkRef(link, doFn); err != nil {
			return err
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

	for _, pathItem := range *ref.Value {
		for _, parameter := range pathItem.Parameters {
			if err := walkParameterRef(parameter, doFn); err != nil {
				return err
			}
		}
		if err := walkOperations(doFn,
			pathItem.Connect,
			pathItem.Delete,
			pathItem.Get,
			pathItem.Head,
			pathItem.Options,
			pathItem.Patch,
			pathItem.Post,
			pathItem.Put,
			pathItem.Trace,
		); err != nil {
			return err
		}
	}

	return nil
}

// walkOperations calls walkOperation with doFn for each operation.
func walkOperations(doFn func(RefWrapper) (bool, error), operations ...*openapi3.Operation) error {
	for _, op := range operations {
		if err := walkOperation(op, doFn); err != nil {
			return err
		}
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

	return walkSchemaRef(ref.Value.Schema, doFn)
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

func findComponentRefs(swagger *openapi3.T) ([]string, error) {
	refs := []string{}

	if err := walkSwagger(swagger, func(ref RefWrapper) (bool, error) {
		if ref.Ref != "" {
			refs = append(refs, ref.Ref)
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	return refs, nil
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

func pruneUnusedComponents(swagger *openapi3.T) error {
	for {
		refs, err := findComponentRefs(swagger)
		if err != nil {
			return err
		}
		countRemoved := removeOrphanedComponents(swagger, refs)
		if countRemoved < 1 {
			break
		}
	}

	return nil
}
