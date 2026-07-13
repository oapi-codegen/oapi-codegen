// Package routeorder is a regression test for issue-1887: server route
// registration is emitted in the order paths are declared in the spec, not
// sorted. On order-dependent routers (Fiber) this lets the user control
// precedence between overlapping paths — here /templates/{visibility}/shortcuts
// is declared before /templates/privates/{id} so it is matched first.
package routeorder

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
