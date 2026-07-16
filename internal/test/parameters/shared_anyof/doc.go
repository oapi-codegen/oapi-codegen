// Package sharedanyof is a regression test for issue-2090: a path-level anyOf
// parameter shared by multiple methods on a path (and on webhooks/callbacks)
// must have its union member types declared once for the path item, not once
// per method. The committed generated file compiling as part of the test
// module is the guard.
package sharedanyof

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
