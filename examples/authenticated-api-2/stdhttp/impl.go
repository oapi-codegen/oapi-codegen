package stdhttp

import "net/http"

type Server struct{}

// (GET /apiKey)
func (*Server) ApiKey(w http.ResponseWriter, r *http.Request) {}

// (GET /httpBasic)
func (*Server) HttpBasic(w http.ResponseWriter, r *http.Request) {}

// (GET /unauthenticated)
func (*Server) Unauthenticated(w http.ResponseWriter, r *http.Request) {}
