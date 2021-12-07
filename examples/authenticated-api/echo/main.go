package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/deepmap/oapi-codegen/examples/authenticated-api/echo/api"
	"github.com/deepmap/oapi-codegen/examples/authenticated-api/echo/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	var port = flag.Int("port", 8080, "port where to serve traffic")

	e := echo.New()

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
	e.Use(middleware.Logger())
	e.Use(mw...)

	svr := server.NewServer()

	api.RegisterHandlers(e, svr)

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

	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", *port)))
}
