// Package parameterssharedcollision is a regression fixture for the naming of
// helper types hoisted by shared (path-item-level) parameters (issue #2090).
//
// A parameter declared at the path-item level is inherited by every method on
// the path, so a helper type it hoists (e.g. an anyOf member) was declared once
// per operation and redeclared — "Id0 redeclared in this block" — on any path
// with more than one method, and clashed again across sibling paths that reuse
// the same parameter. The spec exercises each case: a parameter shared across
// two methods of one path (declared once, bare), the same parameter reused by
// two sibling paths (hash-disambiguated per path), a non-colliding single-method
// parameter (unchanged), and the equivalent shared parameters on a webhook and a
// callback (the emit-once paths in WebhookOperationDefinitions and
// CallbackOperationDefinitions).
//
// The generated file compiling as part of this package is the primary guard:
// the redeclaration bug would fail the build.
package parameterssharedcollision

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
