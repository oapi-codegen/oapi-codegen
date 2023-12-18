package codegen

import (
	"fmt"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
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
	SourceRef interface{}
}

func walkSwagger(swagger *libopenapi.DocumentModel[v3.Document], doFn func(RefWrapper) (bool, error)) error {
	if swagger == nil || swagger.Model.Paths == nil {
		return nil
	}

	for _, p := range ToMap(swagger.Model.Paths.PathItems) {
		for _, param := range p.Parameters {
			_ = walkParameterRef(param, doFn)
		}
		for _, op := range ToMap(p.GetOperations()) {
			_ = walkOperation(op, doFn)
		}
	}

	_ = walkComponents(swagger.Model.Components, doFn)

	return nil
}

func walkOperation(op *v3.Operation, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if op == nil {
		return nil
	}

	for _, param := range op.Parameters {
		_ = walkParameterRef(param, doFn)
	}

	_ = walkRequestBodyRef(op.RequestBody, doFn)

	if op.Responses != nil {
		for _, response := range ToMap(GetResponsesWithDefault(op.Responses)) {
			_ = walkResponseRef(response, doFn)
		}
	}

	for _, callback := range ToMap(op.Callbacks) {
		_ = walkCallbackRef(callback, doFn)
	}

	return nil
}

func walkComponents(components *v3.Components, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if components == nil {
		return nil
	}

	for _, schema := range ToMap(components.Schemas) {
		_ = walkSchemaRef(schema, doFn)
	}

	for _, param := range ToMap(components.Parameters) {
		_ = walkParameterRef(param, doFn)
	}

	for _, header := range ToMap(components.Headers) {
		_ = walkHeaderRef(header, doFn)
	}

	for _, requestBody := range ToMap(components.RequestBodies) {
		_ = walkRequestBodyRef(requestBody, doFn)
	}

	for _, response := range ToMap(components.Responses) {
		_ = walkResponseRef(response, doFn)
	}

	for _, securityScheme := range ToMap(components.SecuritySchemes) {
		_ = walkSecuritySchemeRef(securityScheme, doFn)
	}

	for _, example := range ToMap(components.Examples) {
		_ = walkExampleRef(example, doFn)
	}

	for _, link := range ToMap(components.Links) {
		_ = walkLinkRef(link, doFn)
	}

	for _, callback := range ToMap(components.Callbacks) {
		_ = walkCallbackRef(callback, doFn)
	}

	return nil
}

func walkSchemaRef(ref *base.SchemaProxy, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	for _, ref := range ref.Schema().OneOf {
		_ = walkSchemaRef(ref, doFn)
	}

	for _, ref := range ref.Schema().AnyOf {
		_ = walkSchemaRef(ref, doFn)
	}

	for _, ref := range ref.Schema().AllOf {
		_ = walkSchemaRef(ref, doFn)
	}

	_ = walkSchemaRef(ref.Schema().Not, doFn)

	// With OpenAPI 3.1 Items can be a boolean, in that case there is no schema to walk
	if ref.Schema().Items != nil && ref.Schema().Items.IsA() {
		_ = walkSchemaRef(ref.Schema().Items.A, doFn)
	}

	for _, ref := range ToMap(ref.Schema().Properties) {
		_ = walkSchemaRef(ref, doFn)
	}

	if ref.Schema().AdditionalProperties != nil && ref.Schema().AdditionalProperties.IsA() {
		_ = walkSchemaRef(ref.Schema().AdditionalProperties.A, doFn)
	}
	return nil
}

func walkParameterRef(ref *v3.Parameter, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GoLow().GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	_ = walkSchemaRef(ref.Schema, doFn)

	for _, example := range ToMap(ref.Examples) {
		_ = walkExampleRef(example, doFn)
	}

	for _, mediaType := range ToMap(ref.Content) {
		if mediaType == nil {
			continue
		}
		_ = walkSchemaRef(mediaType.Schema, doFn)

		for _, example := range ToMap(mediaType.Examples) {
			_ = walkExampleRef(example, doFn)
		}
	}

	return nil
}

func walkRequestBodyRef(ref *v3.RequestBody, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GoLow().GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	for _, mediaType := range ToMap(ref.Content) {
		if mediaType == nil {
			continue
		}
		_ = walkSchemaRef(mediaType.Schema, doFn)

		for _, example := range ToMap(mediaType.Examples) {
			_ = walkExampleRef(example, doFn)
		}
	}

	return nil
}

func walkResponseRef(ref *v3.Response, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GoLow().GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	for _, header := range ToMap(ref.Headers) {
		_ = walkHeaderRef(header, doFn)
	}

	for _, mediaType := range ToMap(ref.Content) {
		if mediaType == nil {
			continue
		}
		_ = walkSchemaRef(mediaType.Schema, doFn)

		for _, example := range ToMap(mediaType.Examples) {
			_ = walkExampleRef(example, doFn)
		}
	}

	for _, link := range ToMap(ref.Links) {
		_ = walkLinkRef(link, doFn)
	}

	return nil
}

func walkCallbackRef(ref *v3.Callback, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GoLow().GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	for _, pathItem := range ToMap(ref.Expression) {
		for _, parameter := range pathItem.Parameters {
			_ = walkParameterRef(parameter, doFn)
		}
		// _ = walkOperation(pathItem.Connect, doFn)
		_ = walkOperation(pathItem.Delete, doFn)
		_ = walkOperation(pathItem.Get, doFn)
		_ = walkOperation(pathItem.Head, doFn)
		_ = walkOperation(pathItem.Options, doFn)
		_ = walkOperation(pathItem.Patch, doFn)
		_ = walkOperation(pathItem.Post, doFn)
		_ = walkOperation(pathItem.Put, doFn)
		_ = walkOperation(pathItem.Trace, doFn)
	}

	return nil
}

func walkHeaderRef(ref *v3.Header, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GoLow().GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	_ = walkSchemaRef(ref.Schema, doFn)

	return nil
}

func walkSecuritySchemeRef(ref *v3.SecurityScheme, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GoLow().GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	// NOTE: `SecuritySchemeRef`s don't contain any children that can contain refs

	return nil
}

func walkLinkRef(ref *v3.Link, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GoLow().GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	return nil
}

func walkExampleRef(ref *base.Example, doFn func(RefWrapper) (bool, error)) error {
	// Not a valid ref, ignore it and continue
	if ref == nil {
		return nil
	}
	refWrapper := RefWrapper{Ref: ref.GoLow().GetReference(), SourceRef: ref}
	shouldContinue, err := doFn(refWrapper)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	// NOTE: `ExampleRef`s don't contain any children that can contain refs

	return nil
}

func findComponentRefs(swagger *libopenapi.DocumentModel[v3.Document]) []string {
	refs := []string{}

	_ = walkSwagger(swagger, func(ref RefWrapper) (bool, error) {
		if ref.Ref != "" {
			refs = append(refs, ref.Ref)
			return false, nil
		}
		return true, nil
	})

	return refs
}

func removeOrphanedComponents(swagger *libopenapi.DocumentModel[v3.Document], refs []string) int {
	if swagger.Model.Components == nil {
		return 0
	}

	countRemoved := 0

	for key := range ToMap(swagger.Model.Components.Schemas) {
		ref := fmt.Sprintf("#/components/schemas/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			swagger.Model.Components.Schemas.Delete(key)
		}
	}

	for key := range ToMap(swagger.Model.Components.Parameters) {
		ref := fmt.Sprintf("#/components/parameters/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			swagger.Model.Components.Parameters.Delete(key)
		}
	}

	// securitySchemes are an exception. definitions in securitySchemes
	// are referenced directly by name. and not by $ref

	// for key, _ := range swagger.Model.Components.SecuritySchemes {
	// 	ref := fmt.Sprintf("#/components/securitySchemes/%s", key)
	// 	if !stringInSlice(ref, refs) {
	// 		countRemoved++
	// 		delete(swagger.Model.Components.SecuritySchemes, key)
	// 	}
	// }

	for key := range ToMap(swagger.Model.Components.RequestBodies) {
		ref := fmt.Sprintf("#/components/requestBodies/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			swagger.Model.Components.RequestBodies.Delete(key)
		}
	}

	for key := range ToMap(swagger.Model.Components.Responses) {
		ref := fmt.Sprintf("#/components/responses/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			swagger.Model.Components.Responses.Delete(key)
		}
	}

	for key := range ToMap(swagger.Model.Components.Headers) {
		ref := fmt.Sprintf("#/components/headers/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			swagger.Model.Components.Headers.Delete(key)
		}
	}

	for key := range ToMap(swagger.Model.Components.Examples) {
		ref := fmt.Sprintf("#/components/examples/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			swagger.Model.Components.Examples.Delete(key)
		}
	}

	for key := range ToMap(swagger.Model.Components.Links) {
		ref := fmt.Sprintf("#/components/links/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			swagger.Model.Components.Links.Delete(key)
		}
	}

	for key := range ToMap(swagger.Model.Components.Callbacks) {
		ref := fmt.Sprintf("#/components/callbacks/%s", key)
		if !stringInSlice(ref, refs) {
			countRemoved++
			swagger.Model.Components.Callbacks.Delete(key)
		}
	}

	return countRemoved
}

func pruneUnusedComponents(swagger *libopenapi.DocumentModel[v3.Document]) {
	for {
		refs := findComponentRefs(swagger)
		countRemoved := removeOrphanedComponents(swagger, refs)
		if countRemoved < 1 {
			break
		}
	}
}
