//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../parameters.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../parameters.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=client.cfg.yaml ../parameters.yaml

package echov5params

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type Server struct{}

func (s *Server) GetContentObject(ctx *echo.Context, param ComplexObject) error { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetCookie(ctx *echo.Context, params GetCookieParams) error     { return (*ctx).JSON(http.StatusOK, params) }
func (s *Server) EnumParams(ctx *echo.Context, params EnumParamsParams) error   { return (*ctx).NoContent(http.StatusNoContent) }
func (s *Server) GetHeader(ctx *echo.Context, params GetHeaderParams) error     { return (*ctx).JSON(http.StatusOK, params) }
func (s *Server) GetLabelExplodeArray(ctx *echo.Context, param []int32) error   { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetLabelExplodeObject(ctx *echo.Context, param Object) error   { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetLabelExplodePrimitive(ctx *echo.Context, param int32) error { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetLabelNoExplodeArray(ctx *echo.Context, param []int32) error { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetLabelNoExplodeObject(ctx *echo.Context, param Object) error { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetLabelPrimitive(ctx *echo.Context, param int32) error        { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetMatrixExplodeArray(ctx *echo.Context, id []int32) error     { return (*ctx).JSON(http.StatusOK, id) }
func (s *Server) GetMatrixExplodeObject(ctx *echo.Context, id Object) error     { return (*ctx).JSON(http.StatusOK, id) }
func (s *Server) GetMatrixExplodePrimitive(ctx *echo.Context, id int32) error   { return (*ctx).JSON(http.StatusOK, id) }
func (s *Server) GetMatrixNoExplodeArray(ctx *echo.Context, id []int32) error   { return (*ctx).JSON(http.StatusOK, id) }
func (s *Server) GetMatrixNoExplodeObject(ctx *echo.Context, id Object) error   { return (*ctx).JSON(http.StatusOK, id) }
func (s *Server) GetMatrixPrimitive(ctx *echo.Context, id int32) error          { return (*ctx).JSON(http.StatusOK, id) }
func (s *Server) GetPassThrough(ctx *echo.Context, param string) error          { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetDeepObject(ctx *echo.Context, params GetDeepObjectParams) error { return (*ctx).JSON(http.StatusOK, params) }
func (s *Server) GetQueryDelimited(ctx *echo.Context, params GetQueryDelimitedParams) error { return (*ctx).JSON(http.StatusOK, params) }
func (s *Server) GetQueryForm(ctx *echo.Context, params GetQueryFormParams) error { return (*ctx).JSON(http.StatusOK, params) }
func (s *Server) GetSimpleExplodeArray(ctx *echo.Context, param []int32) error    { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetSimpleExplodeObject(ctx *echo.Context, param Object) error    { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetSimpleExplodePrimitive(ctx *echo.Context, param int32) error  { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetSimpleNoExplodeArray(ctx *echo.Context, param []int32) error  { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetSimpleNoExplodeObject(ctx *echo.Context, param Object) error  { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetSimplePrimitive(ctx *echo.Context, param int32) error         { return (*ctx).JSON(http.StatusOK, param) }
func (s *Server) GetStartingWithNumber(ctx *echo.Context, n1param string) error   { return (*ctx).JSON(http.StatusOK, n1param) }
