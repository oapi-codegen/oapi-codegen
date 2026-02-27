//go:build tools

// This file exists solely to keep github.com/oapi-codegen/oapi-codegen/v2 in
// go.mod's require block. The go:generate directives in doc.go files use
// "go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen", which
// needs the module to be required — but since no test code imports it directly,
// go mod tidy would otherwise prune it. The replace directive in go.mod
// redirects it to the local checkout (../../), but replace alone is not
// sufficient without a corresponding require.
//
// In Go 1.24+ this file can be replaced by a "tool" directive in go.mod.

package tools

import _ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
