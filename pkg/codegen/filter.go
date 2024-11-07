package codegen

import "github.com/getkin/kin-openapi/openapi3"

func sliceToMap(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}

func filterOperationsByTag(swagger *openapi3.T, opts Configuration) {
	if len(opts.OutputOptions.ExcludeTags) > 0 {
		operationsWithTags(swagger.Paths, sliceToMap(opts.OutputOptions.ExcludeTags), true)
	}
	if len(opts.OutputOptions.IncludeTags) > 0 {
		operationsWithTags(swagger.Paths, sliceToMap(opts.OutputOptions.IncludeTags), false)
	}
}

func operationsWithTags(paths *openapi3.Paths, tags map[string]bool, exclude bool) {
	if paths == nil {
		return
	}

	for _, pathItem := range paths.Map() {
		ops := pathItem.Operations()
		names := make([]string, 0, len(ops))
		for name, op := range ops {
			if operationHasTag(op, tags) == exclude {
				names = append(names, name)
			}
		}
		for _, name := range names {
			pathItem.SetOperation(name, nil)
		}
	}
}

// operationHasTag returns true if the operation is tagged with any of tags
func operationHasTag(op *openapi3.Operation, tags map[string]bool) bool {
	if op == nil {
		return false
	}
	for _, hasTag := range op.Tags {
		if tags[hasTag] {
			return true
		}
	}
	return false
}

func filterOperationsByOperationID(swagger *openapi3.T, opts Configuration) {
	if len(opts.OutputOptions.ExcludeOperationIDs) > 0 {
		operationsWithOperationIDs(swagger.Paths, sliceToMap(opts.OutputOptions.ExcludeOperationIDs), true)
	}
	if len(opts.OutputOptions.IncludeOperationIDs) > 0 {
		operationsWithOperationIDs(swagger.Paths, sliceToMap(opts.OutputOptions.IncludeOperationIDs), false)
	}
}

func operationsWithOperationIDs(paths *openapi3.Paths, operationIDs map[string]bool, exclude bool) {
	if paths == nil {
		return
	}

	for _, pathItem := range paths.Map() {
		ops := pathItem.Operations()
		names := make([]string, 0, len(ops))
		for name, op := range ops {
			if operationHasOperationID(op, operationIDs) == exclude {
				names = append(names, name)
			}
		}
		for _, name := range names {
			pathItem.SetOperation(name, nil)
		}
	}
}

// operationHasOperationID returns true if the operation has operation id is included in operation ids
func operationHasOperationID(op *openapi3.Operation, operationIDs map[string]bool) bool {
	if op == nil {
		return false
	}
	return operationIDs[op.OperationID]
}
