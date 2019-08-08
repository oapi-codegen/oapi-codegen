module github.com/weberr13/oapi-codegen

require (
	github.com/cyberdelia/templates v0.0.0-20141128023046-ca7fffd4298c
	github.com/deepmap/oapi-codegen v1.3.4
	github.com/getkin/kin-openapi v0.2.0
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/golangci/lint-1 v0.0.0-20181222135242-d2cdd8c08219
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/labstack/echo/v4 v4.1.11
	github.com/matryer/moq v0.0.0-20190312154309-6cfb0558e1bd
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	golang.org/x/tools v0.0.0-20200203222849-174f5c63c9f5 // indirect
)

go 1.13

replace github.com/getkin/kin-openapi => github.com/weberr13/kin-openapi v0.2.6
