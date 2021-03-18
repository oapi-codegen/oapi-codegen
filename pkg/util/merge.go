package util

import (
	"fmt"
	"context"
	
	"github.com/getkin/kin-openapi/openapi3"
)

func Merge(ctx context.Context, target *openapi3.Swagger, toMerge *openapi3.Swagger) (err error) {
	for pathName, pathItem := range toMerge.Paths {
		if _, ok := target.Paths[pathName]; ok {
			return fmt.Errorf("path already exists: %s", pathName)
		}
		target.Paths[pathName] = pathItem
	}
	if err = target.Validate(ctx); err != nil {
		return err
	}
	return
}

func MergeFunc(ctx context.Context, target *openapi3.Swagger, toMergeFunc func() (*openapi3.Swagger, error)) (err error) {
	var toMerge *openapi3.Swagger
	if toMerge, err = toMergeFunc(); err != nil {
		return err
	}
	if err = Merge(ctx, target, toMerge); err != nil {
		return err
	}
	return
}

// Merge all openapi spec into a single openapi3.Swagger container.
// This is needed for validating request/response based on openapi specification.
func MergeAll(ctx context.Context, toMergeFuncList ...func() (*openapi3.Swagger, error)) (result *openapi3.Swagger, err error) {
	for _, toMergeFunc := range toMergeFuncList {
		var toMerge *openapi3.Swagger
		if toMerge, err = toMergeFunc(); err != nil {
			return
		}
		if err = toMerge.Validate(ctx); err != nil {
			return
		}

		if result == nil {
			result = toMerge
			continue
		}

		if err = Merge(ctx, result, toMerge); err != nil {
			return
		}
	}
	if result == nil {
		result = &openapi3.Swagger{}
	}
	return
}
