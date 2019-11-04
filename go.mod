module github.com/lukehamilton/oapi-codegen

replace github.com/deepmap/oapi-codegen => ./

require (
	github.com/cyberdelia/templates v0.0.0-20141128023046-ca7fffd4298c
	github.com/deepmap/oapi-codegen v0.0.0-00010101000000-000000000000
	github.com/getkin/kin-openapi v0.2.0
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/golangci/lint-1 v0.0.0-20181222135242-d2cdd8c08219
	github.com/labstack/echo/v4 v4.1.11
	github.com/matryer/moq v0.0.0-20190312154309-6cfb0558e1bd
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	golang.org/x/tools v0.0.0-20191101200257-8dbcdeb83d3f // indirect
)

go 1.13
