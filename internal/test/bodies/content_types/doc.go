// Package bodiescontenttypes exercises request/response content-type handling.
// Most cases are compile-only (a regression would fail to generate/build); the
// text/plain coercion case carries client assertions in the test file.
//
//   - schemas.yaml Issue127:  unsupported content types must not preempt supported ones.
//   - schemas.yaml Issue1051: multiple media types that all contain JSON.
//   - issue #1127:      request/response body with several +json media types.
//   - issue #1168:      application/problem+json responses.
//   - issue-vnd-json:   application/vnd.api+json (incl. anyOf error body).
//   - issue #1914:      text/plain request body coerces UUID/number to string.
//
// Strict-server content-type regressions that need per-framework server codegen
// live in the multi_json and text_and_json sub-packages.
package bodiescontenttypes

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
