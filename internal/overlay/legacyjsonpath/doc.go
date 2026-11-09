// Package legacyjsonpath translates overlay targets written in the JSONPath
// dialect that oapi-codegen accepted before it moved to RFC 9535, into RFC
// 9535 JSONPath.
//
// The previous dialect is the one implemented by
// github.com/vmware-labs/yaml-jsonpath, which oapi-codegen used through
// github.com/speakeasy-api/openapi for Overlay 1.0.0 documents. Translate
// rewrites the constructs that differ:
//
//   - A dot-notation name that RFC 9535 does not allow in shorthand form,
//     such as .x-internal, .application/json or .200, becomes a bracketed
//     name: ['x-internal']. This applies inside filters too: @.x-internal
//     becomes @['x-internal'].
//   - A target that does not start with $ gets $. in front of it.
//   - A wildcard immediately before a filter is dropped: .*[?(P)] and
//     [*][?(P)] become [?P]. The previous dialect applied a filter that
//     follows a mapping to the mapping itself, whereas an RFC 9535 filter
//     tests the children of the node it follows, so the filter has to move
//     up one level. A filter after recursive descent, ..[?(P)], keeps its
//     place.
//   - X =~ /re/ becomes search(X, 're'). Flags after the closing / are an
//     error, as they were before.
//   - A comparison whose query selects several nodes through a wildcard,
//     X[*] == v, becomes a nested filter, X[?@ == v]. RFC 9535 only allows
//     queries that select at most one node in comparisons.
//   - A filter directly after a named node, .name[?(P)], is an error. The
//     previous dialect tested the node itself when it was a mapping and its
//     elements when it was a sequence; RFC 9535 cannot test a node by its own
//     key. Writing .name[*][?(P)] selects the elements of a sequence in
//     both dialects.
//   - The ~ property-name operator is an error: RFC 9535 has no equivalent.
//
// The translation is an approximation of the previous semantics, not an
// emulation. The known gaps are:
//
//   - A wildcard before a filter selected a sequence's elements when the
//     wildcard's node was a sequence: in .*[?(P)] over a mapping whose values
//     include a sequence, the previous dialect tested the sequence's elements
//     while [?P] tests the sequence itself.
//   - A comparison over several nodes, X[*] == v, was true when every node
//     compared true; the nested filter X[?@ == v] is true when any node does.
//     The same holds for != and the ordering operators, and for =~.
//   - Recursive descent included the node it started from, so
//     schema..[?(P)] could select schema itself; in RFC 9535 it selects only
//     schema's descendants.
//
// Targets that use these constructs should be rewritten in RFC 9535 and the
// compatibility option turned off.
package legacyjsonpath
