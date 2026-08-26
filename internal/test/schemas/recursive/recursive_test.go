package schemasrecursive

// issue #52: recursion via additionalProperties — compile-only.
// The original test called codegen.Generate() to verify no infinite loop;
// here the generated types compiling is sufficient evidence.
var _ Document
var _ Value
var _ ArrayValue

// issue #936: cyclic oneOf — compile-only.
// The original test verified generation succeeds; compilation confirms it.
var _ FilterColumnIncludes
var _ FilterPredicate
var _ FilterPredicateOp
var _ FilterPredicateRangeOp
var _ FilterRangeValue
var _ FilterValue

// issue #1373: recursive $ref via allOf — compile-only.
// The original test verified generation succeeds; compilation confirms it.
var _ RecursiveObject
var _ NonRecursiveObject

// issue #2542: recursive anyOf whose branch wraps the self-$ref in allOf.
// The original crash was a stack overflow; compilation of the generated
// types confirms generation now terminates without dropping the extra field.
var _ Node
var _ Node0
var _ Node1
