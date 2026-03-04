package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/authenticated-api/stdhttp/api"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/authenticated-api/stdhttp/server"
)

func main() {
	port := flag.String("port", "8080", "port where to serve traffic")

	r := http.NewServeMux()

	// Create a fake authenticator. This allows us to issue tokens, and also
	// implements a validator to check their validity.
	fa, err := server.NewFakeAuthenticator()
	if err != nil {
		log.Fatalln("error creating authenticator:", err)
	}

	// Create middleware for validating tokens.
	mw, err := server.CreateMiddleware(fa)
	if err != nil {
		log.Fatalln("error creating middleware:", err)
	}

	svr := server.NewServer()

	h := api.HandlerFromMux(svr, r)
	// wrap the existing handler with our global middleware
	h = mw(h)

	// We're going to print some useful things for interacting with this server.
	// This token allows access to any API's with no specific claims.
	readerJWS, err := fa.CreateJWSWithClaims([]string{})
	if err != nil {
		log.Fatalln("error creating reader JWS:", err)
	}
	// This token allows access to API's with no scopes, and with the "things:w" claim.
	writerJWS, err := fa.CreateJWSWithClaims([]string{"things:w"})
	if err != nil {
		log.Fatalln("error creating writer JWS:", err)
	}

	log.Println("Reader token", string(readerJWS))
	log.Println("Writer token", string(writerJWS))

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:" + *port,
	}

	log.Fatal(s.ListenAndServe())
}
