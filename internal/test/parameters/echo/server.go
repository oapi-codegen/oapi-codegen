//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../parameters.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../parameters.yaml

package echoparams

import (
	"net/http"

	"github.com/labstack/echo/v4"

	gen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/echo/gen"
)

type Server struct{}

var _ gen.ServerInterface = (*Server)(nil)

func (s *Server) GetContentObject(ctx echo.Context, param gen.ComplexObject) error              { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetCookie(ctx echo.Context, params gen.GetCookieParams) error                  { return ctx.JSON(http.StatusOK, params) }
func (s *Server) EnumParams(ctx echo.Context, params gen.EnumParamsParams) error                { return ctx.NoContent(http.StatusNoContent) }
func (s *Server) GetHeader(ctx echo.Context, params gen.GetHeaderParams) error                  { return ctx.JSON(http.StatusOK, params) }
func (s *Server) GetLabelExplodeArray(ctx echo.Context, param []int32) error                    { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetLabelExplodeObject(ctx echo.Context, param gen.Object) error                { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetLabelExplodePrimitive(ctx echo.Context, param int32) error                  { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetLabelNoExplodeArray(ctx echo.Context, param []int32) error                  { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetLabelNoExplodeObject(ctx echo.Context, param gen.Object) error              { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetLabelPrimitive(ctx echo.Context, param int32) error                         { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetMatrixExplodeArray(ctx echo.Context, id []int32) error                      { return ctx.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixExplodeObject(ctx echo.Context, id gen.Object) error                  { return ctx.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixExplodePrimitive(ctx echo.Context, id int32) error                    { return ctx.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixNoExplodeArray(ctx echo.Context, id []int32) error                    { return ctx.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixNoExplodeObject(ctx echo.Context, id gen.Object) error                { return ctx.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixPrimitive(ctx echo.Context, id int32) error                           { return ctx.JSON(http.StatusOK, id) }
func (s *Server) GetPassThrough(ctx echo.Context, param string) error                           { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetDeepObject(ctx echo.Context, params gen.GetDeepObjectParams) error          { return ctx.JSON(http.StatusOK, params) }
func (s *Server) GetQueryDelimited(ctx echo.Context, params gen.GetQueryDelimitedParams) error  { return ctx.JSON(http.StatusOK, params) }
func (s *Server) GetQueryForm(ctx echo.Context, params gen.GetQueryFormParams) error            { return ctx.JSON(http.StatusOK, params) }
func (s *Server) GetSimpleExplodeArray(ctx echo.Context, param []int32) error                   { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetSimpleExplodeObject(ctx echo.Context, param gen.Object) error               { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetSimpleExplodePrimitive(ctx echo.Context, param int32) error                 { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetSimpleNoExplodeArray(ctx echo.Context, param []int32) error                 { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetSimpleNoExplodeObject(ctx echo.Context, param gen.Object) error             { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetSimplePrimitive(ctx echo.Context, param int32) error                        { return ctx.JSON(http.StatusOK, param) }
func (s *Server) GetStartingWithNumber(ctx echo.Context, n1param string) error                  { return ctx.JSON(http.StatusOK, n1param) }
