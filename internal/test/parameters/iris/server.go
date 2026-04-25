//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../parameters.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../parameters.yaml

package irisparams

import (
	"net/http"

	"github.com/kataras/iris/v12"

	gen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/iris/gen"
)

type Server struct{}

var _ gen.ServerInterface = (*Server)(nil)

func (s *Server) GetContentObject(ctx iris.Context, param gen.ComplexObject)                { _ = ctx.JSON(param) }
func (s *Server) GetCookie(ctx iris.Context, params gen.GetCookieParams)                    { _ = ctx.JSON(params) }
func (s *Server) EnumParams(ctx iris.Context, params gen.EnumParamsParams)                  { ctx.StatusCode(http.StatusNoContent) }
func (s *Server) GetHeader(ctx iris.Context, params gen.GetHeaderParams)                    { _ = ctx.JSON(params) }
func (s *Server) GetLabelExplodeArray(ctx iris.Context, param []int32)                      { _ = ctx.JSON(param) }
func (s *Server) GetLabelExplodeObject(ctx iris.Context, param gen.Object)                  { _ = ctx.JSON(param) }
func (s *Server) GetLabelExplodePrimitive(ctx iris.Context, param int32)                    { _ = ctx.JSON(param) }
func (s *Server) GetLabelNoExplodeArray(ctx iris.Context, param []int32)                    { _ = ctx.JSON(param) }
func (s *Server) GetLabelNoExplodeObject(ctx iris.Context, param gen.Object)                { _ = ctx.JSON(param) }
func (s *Server) GetLabelPrimitive(ctx iris.Context, param int32)                           { _ = ctx.JSON(param) }
func (s *Server) GetMatrixExplodeArray(ctx iris.Context, id []int32)                        { _ = ctx.JSON(id) }
func (s *Server) GetMatrixExplodeObject(ctx iris.Context, id gen.Object)                    { _ = ctx.JSON(id) }
func (s *Server) GetMatrixExplodePrimitive(ctx iris.Context, id int32)                      { _ = ctx.JSON(id) }
func (s *Server) GetMatrixNoExplodeArray(ctx iris.Context, id []int32)                      { _ = ctx.JSON(id) }
func (s *Server) GetMatrixNoExplodeObject(ctx iris.Context, id gen.Object)                  { _ = ctx.JSON(id) }
func (s *Server) GetMatrixPrimitive(ctx iris.Context, id int32)                             { _ = ctx.JSON(id) }
func (s *Server) GetPassThrough(ctx iris.Context, param string)                             { _ = ctx.JSON(param) }
func (s *Server) GetDeepObject(ctx iris.Context, params gen.GetDeepObjectParams)             { _ = ctx.JSON(params) }
func (s *Server) GetQueryDelimited(ctx iris.Context, params gen.GetQueryDelimitedParams)     { _ = ctx.JSON(params) }
func (s *Server) GetQueryForm(ctx iris.Context, params gen.GetQueryFormParams)               { _ = ctx.JSON(params) }
func (s *Server) GetSimpleExplodeArray(ctx iris.Context, param []int32)                     { _ = ctx.JSON(param) }
func (s *Server) GetSimpleExplodeObject(ctx iris.Context, param gen.Object)                 { _ = ctx.JSON(param) }
func (s *Server) GetSimpleExplodePrimitive(ctx iris.Context, param int32)                   { _ = ctx.JSON(param) }
func (s *Server) GetSimpleNoExplodeArray(ctx iris.Context, param []int32)                   { _ = ctx.JSON(param) }
func (s *Server) GetSimpleNoExplodeObject(ctx iris.Context, param gen.Object)               { _ = ctx.JSON(param) }
func (s *Server) GetSimplePrimitive(ctx iris.Context, param int32)                          { _ = ctx.JSON(param) }
func (s *Server) GetStartingWithNumber(ctx iris.Context, n1param string)                    { _ = ctx.JSON(n1param) }
