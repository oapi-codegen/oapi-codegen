package codegen

// FilterOperationsByTag filters operations based on include/exclude tag lists.
// Exclude is applied first, then include.
func FilterOperationsByTag(ops []*OperationDescriptor, opts OutputOptions) []*OperationDescriptor {
	if len(opts.ExcludeTags) > 0 {
		tags := sliceToSet(opts.ExcludeTags)
		ops = filterOps(ops, func(op *OperationDescriptor) bool {
			return !operationHasTag(op, tags)
		})
	}
	if len(opts.IncludeTags) > 0 {
		tags := sliceToSet(opts.IncludeTags)
		ops = filterOps(ops, func(op *OperationDescriptor) bool {
			return operationHasTag(op, tags)
		})
	}
	return ops
}

// FilterOperationsByOperationID filters operations based on include/exclude operation ID lists.
// Exclude is applied first, then include.
func FilterOperationsByOperationID(ops []*OperationDescriptor, opts OutputOptions) []*OperationDescriptor {
	if len(opts.ExcludeOperationIDs) > 0 {
		ids := sliceToSet(opts.ExcludeOperationIDs)
		ops = filterOps(ops, func(op *OperationDescriptor) bool {
			return !ids[op.OperationID]
		})
	}
	if len(opts.IncludeOperationIDs) > 0 {
		ids := sliceToSet(opts.IncludeOperationIDs)
		ops = filterOps(ops, func(op *OperationDescriptor) bool {
			return ids[op.OperationID]
		})
	}
	return ops
}

// FilterOperations applies all operation filters (tags, operation IDs) from OutputOptions.
func FilterOperations(ops []*OperationDescriptor, opts OutputOptions) []*OperationDescriptor {
	ops = FilterOperationsByTag(ops, opts)
	ops = FilterOperationsByOperationID(ops, opts)
	return ops
}

// FilterSchemasByName removes schemas whose component name is in the exclude list.
// Only filters top-level component schemas (path: components/schemas/<name>).
func FilterSchemasByName(schemas []*SchemaDescriptor, excludeNames []string) []*SchemaDescriptor {
	if len(excludeNames) == 0 {
		return schemas
	}
	excluded := sliceToSet(excludeNames)
	result := make([]*SchemaDescriptor, 0, len(schemas))
	for _, s := range schemas {
		// Check if this is a top-level component schema
		if len(s.Path) == 3 && s.Path[0] == "components" && s.Path[1] == "schemas" {
			if excluded[s.Path[2]] {
				continue
			}
		}
		result = append(result, s)
	}
	return result
}

// operationHasTag returns true if the operation has any of the given tags.
func operationHasTag(op *OperationDescriptor, tags map[string]bool) bool {
	if op == nil || op.Spec == nil {
		return false
	}
	for _, tag := range op.Spec.Tags {
		if tags[tag] {
			return true
		}
	}
	return false
}

// filterOps returns operations that satisfy the predicate.
func filterOps(ops []*OperationDescriptor, keep func(*OperationDescriptor) bool) []*OperationDescriptor {
	result := make([]*OperationDescriptor, 0, len(ops))
	for _, op := range ops {
		if keep(op) {
			result = append(result, op)
		}
	}
	return result
}

// sliceToSet converts a string slice to a set (map[string]bool).
func sliceToSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}
	return m
}
