package runtime

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kataras/iris/v12"
	"github.com/labstack/echo/v4"
)

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/strictmiddleware/iris#StrictIrisHandlerFunc
type StrictIrisHandlerFunc func(ctx iris.Context, request interface{}) (response interface{}, err error)

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/strictmiddleware/iris#StrictIrisMiddlewareFunc
type StrictIrisMiddlewareFunc func(f StrictIrisHandlerFunc, operationID string) StrictIrisHandlerFunc

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/strictmiddleware/echo#StrictEchoHandlerFunc
type StrictEchoHandlerFunc func(ctx echo.Context, request interface{}) (response interface{}, err error)

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/strictmiddleware/echo#StrictEchoMiddlewareFunc
type StrictEchoMiddlewareFunc func(f StrictEchoHandlerFunc, operationID string) StrictEchoHandlerFunc

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/strictmiddleware/nethttp#StrictHttpHandlerFunc
type StrictHttpHandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error)

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/strictmiddleware/nethttp#StrictHttpMiddlewareFunc
type StrictHttpMiddlewareFunc func(f StrictHttpHandlerFunc, operationID string) StrictHttpHandlerFunc

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/strictmiddleware/gin#StrictGinHandlerFunc
type StrictGinHandlerFunc func(ctx *gin.Context, request interface{}) (response interface{}, err error)

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/strictmiddleware/gin#StrictGinMiddlewareFunc
type StrictGinMiddlewareFunc func(f StrictGinHandlerFunc, operationID string) StrictGinHandlerFunc
