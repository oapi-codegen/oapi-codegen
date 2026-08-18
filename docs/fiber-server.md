# Fiber Server

For a Fiber server, you will want a configuration file such as:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/v2.8.0/configuration-schema.json
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

	"github.com/wayleadr/oapi-codegen/v2/examples/minimal-server/fiber/api"
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

## Header-based API versioning

Alongside `RegisterHandlers`, the Fiber generator emits a `RegisterHandlerVersions`
pair that registers the operations into a shared map keyed by version-prefixed
path instead of binding them on the router directly:

```go
// RegisterHandlerVersions registers the routes into apiHandlers under the given
// version, for dispatch by fibermid.MountVersionedRoutes.
func RegisterHandlerVersions(
	router fiber.Router,
	si ServerInterface,
	version string,
	apiHandlers map[string]map[string]fiber.Handler,
) map[string]map[string]fiber.Handler
```

Each generated package registers into the same map, so several versions of a
spec can coexist. `fibermid.MountVersionedRoutes` then binds the **bare** paths
and dispatches on the `API-Version` (or `X-API-Version`) request header,
falling back to the supplied default version:

```go
import (
	"github.com/gofiber/fiber/v2"

	v1 "example.com/service/api/v1"
	v2 "example.com/service/api/v2"
	"github.com/wayleadr/oapi-codegen/v2/pkg/fibermid"
)

app := fiber.New()

apiHandlers := make(map[string]map[string]fiber.Handler)
v1.RegisterHandlerVersions(app, v1Server, "v1", apiHandlers)
v2.RegisterHandlerVersions(app, v2Server, "v2", apiHandlers)

fibermid.MountVersionedRoutes(app, apiHandlers, "v1")
```

This binds `/events`, `/events/:id` and friends as the only externally visible
URLs — `/v1/...` and `/v2/...` are never registered as routes. Because the
bare-path route does the matching, path parameters resolve normally: the
dispatcher forwards the same `*fiber.Ctx` to the per-version wrapper, so
`c.Params("id")` reads the value the bare-path route extracted.

Each version prefix is also registered under `fibermid.LatestVersion`
(`"latest"`), which is what an unknown or unmatched version falls back to.

> [!NOTE]
> `pkg/fibermid` targets Fiber v2 only. Fiber v3 passes `fiber.Ctx` by value, so
> `fiber-v3-server` does not generate the versioned registration helpers.
