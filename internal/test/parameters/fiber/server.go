//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../parameters.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../parameters.yaml

package fiberparams

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	gen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/fiber/gen"
)

type Server struct{}

var _ gen.ServerInterface = (*Server)(nil)

func (s *Server) GetContentObject(c *fiber.Ctx, param gen.ComplexObject) error                { return c.JSON(param) }
func (s *Server) GetCookie(c *fiber.Ctx, params gen.GetCookieParams) error                    { return c.JSON(params) }
func (s *Server) EnumParams(c *fiber.Ctx, params gen.EnumParamsParams) error                  { return c.SendStatus(http.StatusNoContent) }
func (s *Server) GetHeader(c *fiber.Ctx, params gen.GetHeaderParams) error                    { return c.JSON(params) }
func (s *Server) GetLabelExplodeArray(c *fiber.Ctx, param []int32) error                      { return c.JSON(param) }
func (s *Server) GetLabelExplodeObject(c *fiber.Ctx, param gen.Object) error                  { return c.JSON(param) }
func (s *Server) GetLabelExplodePrimitive(c *fiber.Ctx, param int32) error                    { return c.JSON(param) }
func (s *Server) GetLabelNoExplodeArray(c *fiber.Ctx, param []int32) error                    { return c.JSON(param) }
func (s *Server) GetLabelNoExplodeObject(c *fiber.Ctx, param gen.Object) error                { return c.JSON(param) }
func (s *Server) GetLabelPrimitive(c *fiber.Ctx, param int32) error                           { return c.JSON(param) }
func (s *Server) GetMatrixExplodeArray(c *fiber.Ctx, id []int32) error                        { return c.JSON(id) }
func (s *Server) GetMatrixExplodeObject(c *fiber.Ctx, id gen.Object) error                    { return c.JSON(id) }
func (s *Server) GetMatrixExplodePrimitive(c *fiber.Ctx, id int32) error                      { return c.JSON(id) }
func (s *Server) GetMatrixNoExplodeArray(c *fiber.Ctx, id []int32) error                      { return c.JSON(id) }
func (s *Server) GetMatrixNoExplodeObject(c *fiber.Ctx, id gen.Object) error                  { return c.JSON(id) }
func (s *Server) GetMatrixPrimitive(c *fiber.Ctx, id int32) error                             { return c.JSON(id) }
func (s *Server) GetPassThrough(c *fiber.Ctx, param string) error                             { return c.JSON(param) }
func (s *Server) GetDeepObject(c *fiber.Ctx, params gen.GetDeepObjectParams) error             { return c.JSON(params) }
func (s *Server) GetQueryDelimited(c *fiber.Ctx, params gen.GetQueryDelimitedParams) error     { return c.JSON(params) }
func (s *Server) GetQueryForm(c *fiber.Ctx, params gen.GetQueryFormParams) error               { return c.JSON(params) }
func (s *Server) GetSimpleExplodeArray(c *fiber.Ctx, param []int32) error                     { return c.JSON(param) }
func (s *Server) GetSimpleExplodeObject(c *fiber.Ctx, param gen.Object) error                 { return c.JSON(param) }
func (s *Server) GetSimpleExplodePrimitive(c *fiber.Ctx, param int32) error                   { return c.JSON(param) }
func (s *Server) GetSimpleNoExplodeArray(c *fiber.Ctx, param []int32) error                   { return c.JSON(param) }
func (s *Server) GetSimpleNoExplodeObject(c *fiber.Ctx, param gen.Object) error               { return c.JSON(param) }
func (s *Server) GetSimplePrimitive(c *fiber.Ctx, param int32) error                          { return c.JSON(param) }
func (s *Server) GetStartingWithNumber(c *fiber.Ctx, n1param string) error                    { return c.JSON(n1param) }
