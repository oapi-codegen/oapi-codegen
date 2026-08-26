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
// types confirms generation now terminates without dropping the extra field
// or the reference back to the recursive type (the item keeps a union
// branch referencing Node, matching how a bare $ref terminates recursion).
var _ Node
var _ Node0
var _ Node1

// issue #2542 (object variant): plain recursive object + allOf. Also used
// to overflow the stack; the generated item keeps Extra and a NodeObject
// union reference.
var _ NodeObject
var _ NodeObject_Children_Item

// issue #2542 (transitive variant): the self-$ref hides behind an allOf
// nested inside another allOf member, which escaped the cycle guard (PR
// #2543 review) and hung generation indefinitely.
var _ NodeNestedAllOf
var _ NodeNestedAllOf0
var _ NodeNestedAllOf1
var _ NodeNestedAllOf_1_Children_Item
