//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../parameters.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../parameters.yaml

package chiparams

import (
	"encoding/json"
	"net/http"

	gen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/chi/gen"
)

type Server struct{}

var _ gen.ServerInterface = (*Server)(nil)

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) GetContentObject(w http.ResponseWriter, r *http.Request, param gen.ComplexObject) { writeJSON(w, param) }
func (s *Server) GetCookie(w http.ResponseWriter, r *http.Request, params gen.GetCookieParams) { writeJSON(w, params) }
func (s *Server) EnumParams(w http.ResponseWriter, r *http.Request, params gen.EnumParamsParams) { w.WriteHeader(http.StatusNoContent) }
func (s *Server) GetHeader(w http.ResponseWriter, r *http.Request, params gen.GetHeaderParams) { writeJSON(w, params) }
func (s *Server) GetLabelExplodeArray(w http.ResponseWriter, r *http.Request, param []int32) { writeJSON(w, param) }
func (s *Server) GetLabelExplodeObject(w http.ResponseWriter, r *http.Request, param gen.Object) { writeJSON(w, param) }
func (s *Server) GetLabelExplodePrimitive(w http.ResponseWriter, r *http.Request, param int32) { writeJSON(w, param) }
func (s *Server) GetLabelNoExplodeArray(w http.ResponseWriter, r *http.Request, param []int32) { writeJSON(w, param) }
func (s *Server) GetLabelNoExplodeObject(w http.ResponseWriter, r *http.Request, param gen.Object) { writeJSON(w, param) }
func (s *Server) GetLabelPrimitive(w http.ResponseWriter, r *http.Request, param int32) { writeJSON(w, param) }
func (s *Server) GetMatrixExplodeArray(w http.ResponseWriter, r *http.Request, id []int32) { writeJSON(w, id) }
func (s *Server) GetMatrixExplodeObject(w http.ResponseWriter, r *http.Request, id gen.Object) { writeJSON(w, id) }
func (s *Server) GetMatrixExplodePrimitive(w http.ResponseWriter, r *http.Request, id int32) { writeJSON(w, id) }
func (s *Server) GetMatrixNoExplodeArray(w http.ResponseWriter, r *http.Request, id []int32) { writeJSON(w, id) }
func (s *Server) GetMatrixNoExplodeObject(w http.ResponseWriter, r *http.Request, id gen.Object) { writeJSON(w, id) }
func (s *Server) GetMatrixPrimitive(w http.ResponseWriter, r *http.Request, id int32) { writeJSON(w, id) }
func (s *Server) GetPassThrough(w http.ResponseWriter, r *http.Request, param string) { writeJSON(w, param) }
func (s *Server) GetDeepObject(w http.ResponseWriter, r *http.Request, params gen.GetDeepObjectParams) { writeJSON(w, params) }
func (s *Server) GetQueryDelimited(w http.ResponseWriter, r *http.Request, params gen.GetQueryDelimitedParams) { writeJSON(w, params) }
func (s *Server) GetQueryForm(w http.ResponseWriter, r *http.Request, params gen.GetQueryFormParams) { writeJSON(w, params) }
func (s *Server) GetSimpleExplodeArray(w http.ResponseWriter, r *http.Request, param []int32) { writeJSON(w, param) }
func (s *Server) GetSimpleExplodeObject(w http.ResponseWriter, r *http.Request, param gen.Object) { writeJSON(w, param) }
func (s *Server) GetSimpleExplodePrimitive(w http.ResponseWriter, r *http.Request, param int32) { writeJSON(w, param) }
func (s *Server) GetSimpleNoExplodeArray(w http.ResponseWriter, r *http.Request, param []int32) { writeJSON(w, param) }
func (s *Server) GetSimpleNoExplodeObject(w http.ResponseWriter, r *http.Request, param gen.Object) { writeJSON(w, param) }
func (s *Server) GetSimplePrimitive(w http.ResponseWriter, r *http.Request, param int32) { writeJSON(w, param) }
func (s *Server) GetStartingWithNumber(w http.ResponseWriter, r *http.Request, n1param string) { writeJSON(w, n1param) }
