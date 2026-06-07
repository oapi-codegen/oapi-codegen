# Fiber Server

For a Fiber server, you will want a configuration file such as:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: api
generate:
  fiber-server: true
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
	GetPing(c *fiber.Ctx) error
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router fiber.Router, si ServerInterface) {
	RegisterHandlersWithOptions(router, si, FiberServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router fiber.Router, si ServerInterface, options FiberServerOptions) {
	// ...

	router.Get(options.BaseURL+"/ping", wrapper.GetPing)
}
```

To implement this HTTP server, we need to write the following code in our [`api/impl.go`](../examples/minimal-server/fiber/api/impl.go):

```go
import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ ServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}
}

// (GET /ping)
func (Server) GetPing(ctx *fiber.Ctx) error {
	resp := Pong{
		Ping: "pong",
	}

	return ctx.
		Status(http.StatusOK).
		JSON(resp)
}
```

Now we've got our implementation, we can then write the following code to wire it up and get a running server:

```go
import (
	"log"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/minimal-server/fiber/api"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()

	app := fiber.New()

	api.RegisterHandlers(app, server)

	// And we serve HTTP until the world ends.
	log.Fatal(app.Listen("0.0.0.0:8080"))
}
```

> [!NOTE]
> This doesn't include [validation of incoming requests](../README.md#requestresponse-validation-middleware).
