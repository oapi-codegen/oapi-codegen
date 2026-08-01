// Package serversrouterstrailingslash exercises stdhttp route registration
// for OpenAPI paths with trailing slashes: they must register as
// {$}-anchored exact-match patterns, both to honor OpenAPI's exact-path
// semantics and to avoid ServeMux registration panics when subtree
// patterns from independent paths overlap ambiguously (issue #2065).
package serversrouterstrailingslash

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
