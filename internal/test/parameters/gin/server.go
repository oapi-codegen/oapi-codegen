//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../parameters.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../parameters.yaml

package ginparams

import (
	"net/http"

	"github.com/gin-gonic/gin"

	gen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/gin/gen"
)

type Server struct{}

var _ gen.ServerInterface = (*Server)(nil)

func (s *Server) GetContentObject(c *gin.Context, param gen.ComplexObject)                { c.JSON(http.StatusOK, param) }
func (s *Server) GetCookie(c *gin.Context, params gen.GetCookieParams)                    { c.JSON(http.StatusOK, params) }
func (s *Server) EnumParams(c *gin.Context, params gen.EnumParamsParams)                  { c.Status(http.StatusNoContent) }
func (s *Server) GetHeader(c *gin.Context, params gen.GetHeaderParams)                    { c.JSON(http.StatusOK, params) }
func (s *Server) GetLabelExplodeArray(c *gin.Context, param []int32)                      { c.JSON(http.StatusOK, param) }
func (s *Server) GetLabelExplodeObject(c *gin.Context, param gen.Object)                  { c.JSON(http.StatusOK, param) }
func (s *Server) GetLabelExplodePrimitive(c *gin.Context, param int32)                    { c.JSON(http.StatusOK, param) }
func (s *Server) GetLabelNoExplodeArray(c *gin.Context, param []int32)                    { c.JSON(http.StatusOK, param) }
func (s *Server) GetLabelNoExplodeObject(c *gin.Context, param gen.Object)                { c.JSON(http.StatusOK, param) }
func (s *Server) GetLabelPrimitive(c *gin.Context, param int32)                           { c.JSON(http.StatusOK, param) }
func (s *Server) GetMatrixExplodeArray(c *gin.Context, id []int32)                        { c.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixExplodeObject(c *gin.Context, id gen.Object)                    { c.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixExplodePrimitive(c *gin.Context, id int32)                      { c.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixNoExplodeArray(c *gin.Context, id []int32)                      { c.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixNoExplodeObject(c *gin.Context, id gen.Object)                  { c.JSON(http.StatusOK, id) }
func (s *Server) GetMatrixPrimitive(c *gin.Context, id int32)                             { c.JSON(http.StatusOK, id) }
func (s *Server) GetPassThrough(c *gin.Context, param string)                             { c.JSON(http.StatusOK, param) }
func (s *Server) GetDeepObject(c *gin.Context, params gen.GetDeepObjectParams)             { c.JSON(http.StatusOK, params) }
func (s *Server) GetQueryDelimited(c *gin.Context, params gen.GetQueryDelimitedParams)     { c.JSON(http.StatusOK, params) }
func (s *Server) GetQueryForm(c *gin.Context, params gen.GetQueryFormParams)               { c.JSON(http.StatusOK, params) }
func (s *Server) GetSimpleExplodeArray(c *gin.Context, param []int32)                     { c.JSON(http.StatusOK, param) }
func (s *Server) GetSimpleExplodeObject(c *gin.Context, param gen.Object)                 { c.JSON(http.StatusOK, param) }
func (s *Server) GetSimpleExplodePrimitive(c *gin.Context, param int32)                   { c.JSON(http.StatusOK, param) }
func (s *Server) GetSimpleNoExplodeArray(c *gin.Context, param []int32)                   { c.JSON(http.StatusOK, param) }
func (s *Server) GetSimpleNoExplodeObject(c *gin.Context, param gen.Object)               { c.JSON(http.StatusOK, param) }
func (s *Server) GetSimplePrimitive(c *gin.Context, param int32)                          { c.JSON(http.StatusOK, param) }
func (s *Server) GetStartingWithNumber(c *gin.Context, n1param string)                    { c.JSON(http.StatusOK, n1param) }
