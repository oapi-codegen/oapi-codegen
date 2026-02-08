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

// PruneUnreferencedSchemas removes component schemas that are not $ref'd by any
// other gathered schema. This walks the entire schema descriptor tree, collects
// all $ref paths, and removes component schemas whose path doesn't appear in
// that set. Non-component schemas (inline path schemas, etc.) are always kept.
func PruneUnreferencedSchemas(schemas []*SchemaDescriptor) []*SchemaDescriptor {
	// Collect all $ref paths from all schemas
	referenced := make(map[string]bool)
	for _, s := range schemas {
		collectRefsFromDescriptor(s, referenced)
	}

	result := make([]*SchemaDescriptor, 0, len(schemas))
	for _, s := range schemas {
		// Always keep non-component schemas (inline path schemas, etc.)
		if !s.IsComponentSchema() {
			result = append(result, s)
			continue
		}

		// Keep component schemas that are referenced by something
		if referenced[s.Path.String()] {
			result = append(result, s)
		}
	}
	return result
}

// collectRefsFromDescriptor walks a schema descriptor tree and adds all
// internal $ref paths to the referenced set.
func collectRefsFromDescriptor(desc *SchemaDescriptor, referenced map[string]bool) {
	if desc == nil {
		return
	}

	// If this descriptor has a $ref, record it
	if desc.Ref != "" {
		referenced[desc.Ref] = true
	}

	// Walk children
	for _, child := range desc.Properties {
		collectRefsFromDescriptor(child, referenced)
	}
	if desc.Items != nil {
		collectRefsFromDescriptor(desc.Items, referenced)
	}
	for _, child := range desc.AllOf {
		collectRefsFromDescriptor(child, referenced)
	}
	for _, child := range desc.AnyOf {
		collectRefsFromDescriptor(child, referenced)
	}
	for _, child := range desc.OneOf {
		collectRefsFromDescriptor(child, referenced)
	}
	if desc.AdditionalProps != nil {
		collectRefsFromDescriptor(desc.AdditionalProps, referenced)
	}
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
