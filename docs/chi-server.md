# Chi Server

For a Chi server, you will want a configuration file such as:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: api
generate:
  chi-server: true
  models: true
output: gen.go
```

## Generated code

For instance, let's take this straightforward specification:

```yaml
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Minimal ping API server
paths:
  /ping:
    get:
      responses:
        '200':
          description: pet response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pong'
components:
  schemas:
    # base types
    Pong:
      type: object
      required:
        - ping
      properties:
        ping:
          type: string
          example: pong
```

This then generates code such as:

```go
// Pong defines model for Pong.
type Pong struct {
	Ping string `json:"ping"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /ping)
	GetPing(w http.ResponseWriter, r *http.Request)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r *mux.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	// ...

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/ping", wrapper.GetPing)
	})

	return r
}
```

To implement this HTTP server, we need to write the following code in our [`api/impl.go`](../examples/minimal-server/chi/api/impl.go):

```go
import (
	"encoding/json"
	"net/http"
)

// optional code omitted

type Server struct{}

func NewServer() Server {
	return Server{}
}

// (GET /ping)
func (Server) GetPing(w http.ResponseWriter, r *http.Request) {
	resp := Pong{
		Ping: "pong",
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
```

Now we've got our implementation, we can then write the following code to wire it up and get a running server:

```go
import (
	"log"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/minimal-server/chi/api"
	"github.com/go-chi/chi/v5"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()

	r := chi.NewMux()

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
```

> [!NOTE]
> This doesn't include [validation of incoming requests](../README.md#requestresponse-validation-middleware).
