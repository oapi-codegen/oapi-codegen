module github.com/oapi-codegen/oapi-codegen-exp/experimental/examples/petstore-expanded/echo

go 1.25.0

require (
	github.com/google/uuid v1.6.0
	github.com/labstack/echo/v5 v5.0.0
	github.com/oapi-codegen/oapi-codegen-exp/experimental v0.0.0
)

replace github.com/oapi-codegen/oapi-codegen-exp/experimental => ../../../
