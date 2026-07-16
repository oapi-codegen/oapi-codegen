// Package paths verifies path-handling edge cases in generated servers. It
// currently checks that literal-colon path segments route independently on the
// ':'-parameter routers (Echo, Gin, Fiber v2/v3) rather than colliding as a
// single path parameter. See issue #1726.
package paths

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=echo/config.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=gin/config.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=fiber/config.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=fiberv3/config.yaml spec.yaml
