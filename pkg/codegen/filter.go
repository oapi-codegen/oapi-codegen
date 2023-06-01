package codegen

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func filterOperationsByTag(swagger *libopenapi.DocumentModel[v3.Document], opts Configuration) {
	if len(opts.OutputOptions.ExcludeTags) > 0 {
		excludeOperationsWithTags(swagger.Model.Paths, opts.OutputOptions.ExcludeTags)
	}
	if len(opts.OutputOptions.IncludeTags) > 0 {
		includeOperationsWithTags(swagger.Paths, opts.OutputOptions.IncludeTags, false)
	}
}

func excludeOperationsWithTags(paths *v3.Paths, tags []string) {
	includeOperationsWithTags(paths, tags, true)
}

func includeOperationsWithTags(paths *v3.Paths, tags []string, exclude bool) {
	for _, pathItem := range paths.PathItems {
		ops := pathItem.GetOperations()
		names := make([]string, 0, len(ops))
		for name, op := range ops {
			if operationHasTag(op, tags) == exclude {
				names = append(names, name)
			}
		}
		for _, name := range names {
			paths.PathItems[name] = nil // TODO clear?
		}
	}
}

// operationHasTag returns true if the operation is tagged with any of tags
func operationHasTag(op *v3.Operation, tags []string) bool {
	if op == nil {
		return false
	}
	for _, hasTag := range op.Tags {
		for _, wantTag := range tags {
			if hasTag == wantTag {
				return true
			}
		}
	}
	return false
}
