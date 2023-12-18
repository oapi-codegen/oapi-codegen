package codegen

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

func filterOperationsByTag(swagger *libopenapi.DocumentModel[v3.Document], opts Configuration) {
	if len(opts.OutputOptions.ExcludeTags) > 0 {
		excludeOperationsWithTags(swagger.Model.Paths, opts.OutputOptions.ExcludeTags)
	}
	if len(opts.OutputOptions.IncludeTags) > 0 {
		includeOperationsWithTags(swagger.Model.Paths, opts.OutputOptions.IncludeTags, false)
	}
}

func excludeOperationsWithTags(paths *v3.Paths, tags []string) {
	includeOperationsWithTags(paths, tags, true)
}

func includeOperationsWithTags(paths *v3.Paths, tags []string, exclude bool) {
	if paths == nil {
		return
	}

	for _, pathItem := range ToMap(paths.PathItems) {
		ops := pathItem.GetOperations()
		names := make([]string, 0, orderedmap.Len(ops))
		for name, op := range ToMap(ops) {
			if operationHasTag(op, tags) == exclude {
				names = append(names, name)
			}
		}
		for _, name := range names {
			switch name {
			case "get":
				pathItem.Get = nil
			case "put":
				pathItem.Put = nil
			case "post":
				pathItem.Post = nil
			case "delete":
				pathItem.Delete = nil
			case "options":
				pathItem.Options = nil
			case "head":
				pathItem.Head = nil
			case "patch":
				pathItem.Patch = nil
			case "trace":
				pathItem.Trace = nil
			}
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
