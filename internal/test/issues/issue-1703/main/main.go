package main

import (
	"context"
	"log"
	"net/http"

	issue1703 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-1703"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := NewServer()

	r := http.NewServeMux()

	// get an `http.Handler` that we can use
	h := issue1703.HandlerFromMux(issue1703.NewStrictHandler(server, nil), r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}

type Server struct{}

func NewServer() Server {
	return Server{}
}

// (GET /Test)
func (Server) Test(ctx context.Context, request issue1703.TestRequestObject) (issue1703.TestResponseObject, error) {
	return issue1703.Test200JSONResponse{
		"limit": request.Params.Limit,
	}, nil
}
