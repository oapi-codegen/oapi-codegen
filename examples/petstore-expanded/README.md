# Petstore Expanded Example

This example demonstrates [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) generating server stubs for 9 different Go HTTP frameworks from a single [OpenAPI 3.0 spec](petstore-expanded.yaml) (the canonical Petstore Expanded example).

## Directory Structure

```
petstore-expanded/
├── petstore-expanded.yaml          # Shared OpenAPI spec
├── common/
│   ├── generate.go                 # go:generate for shared model types
│   ├── models.cfg.yaml             # Codegen config: models only
│   ├── models/
│   │   └── models.gen.go           # Generated model types (shared by all variants)
│   ├── store/
│   │   └── store.go                # Framework-agnostic CRUD business logic
│   └── client/
│       ├── main.go                 # CLI test client
│       ├── testclient/
│       │   └── testclient.go       # Reusable test client (shared by integration tests)
│       └── openapi/
│           ├── generate.go         # go:generate for client
│           ├── client.cfg.yaml     # Codegen config: client only
│           └── client.gen.go       # Generated HTTP client
├── chi/                            # Chi (net/http compatible)
├── gorilla/                        # Gorilla/mux (net/http compatible)
├── stdhttp/                        # stdlib net/http (Go 1.22+ ServeMux)
├── echo/                           # Echo v4
├── echo-v5/                        # Echo v5 (requires Go 1.25+, separate module)
├── gin/                            # Gin
├── fiber/                          # Fiber
├── iris/                           # Iris
└── strict/                         # Strict server (Chi + typed request/response objects)
```

Each server variant follows the same pattern:
- `api/server.cfg.yaml` — codegen config generating the server interface and embedded spec
- `api/generate.go` — `//go:generate` directive for the server code
- `api/petstore-server.gen.go` — generated server boilerplate
- `server/server.go` — hand-written `ServerInterface` implementation delegating to `common/store`
- `server/setup.go` — factory function that creates a fully configured server/app
- `petstore.go` — `main()` wiring (thin wrapper around `setup.go`)
- `integration/main.go` — integration test program using the shared test client

## Generating Code

From the `examples/` directory (or repository root with `make generate`):

```sh
# Generate shared models and client
cd examples/petstore-expanded/common && go generate ./...

# Generate a specific server variant
cd examples/petstore-expanded/chi/api && go generate ./...
```

For `echo-v5` (separate Go module):

```sh
cd examples/petstore-expanded/echo-v5/api && go generate ./...
```

## Running a Server

```sh
cd examples/petstore-expanded/chi
go run . --port 8080
```

Replace `chi` with any variant: `gorilla`, `stdhttp`, `echo`, `echo-v5`, `gin`, `fiber`, `iris`, `strict`.

## Integration Tests

Each variant has an integration test program that starts the server on a random port, runs all CRUD operations via the generated HTTP client, and shuts down cleanly:

```sh
cd examples/petstore-expanded
go run ./chi/integration/
go run ./gorilla/integration/
go run ./stdhttp/integration/
go run ./strict/integration/
go run ./gin/integration/
go run ./echo/integration/
go run ./fiber/integration/
go run ./iris/integration/
```

For the separate-module variant:

```sh
cd examples/petstore-expanded/echo-v5
go run ./integration/
```

## CLI Test Client

The test client can also be run manually against a server you start in a separate terminal:

```sh
# Terminal 1: start any server variant
cd examples/petstore-expanded/chi && go run .

# Terminal 2: run the test client
cd examples && go run ./petstore-expanded/common/client/ --port 8080
```

The client verifies: add pets, find by ID, 404 on missing pet, list/filter by tag, delete, and empty list after deletion.
