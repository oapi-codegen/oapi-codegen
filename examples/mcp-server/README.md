# MCP Server Example

This example demonstrates how to generate an MCP (Model Context Protocol) server from an OpenAPI specification using oapi-codegen.

## Overview

The Pet Store API is exposed as a set of MCP tools that can be used by MCP clients (like Claude Desktop or other MCP-compatible applications). Each OpenAPI operation becomes an MCP tool with structured input and output schemas.

## Generated Tools

From the `api.yaml` OpenAPI spec, the following MCP tools are generated:

- `ListPets` - List all pets with optional filtering and pagination
- `CreatePet` - Create a new pet
- `GetPet` - Get a pet by ID
- `UpdatePet` - Update an existing pet
- `DeletePet` - Delete a pet by ID

## Project Structure

- `api.yaml` - OpenAPI 3.0 specification
- `cfg.yaml` - oapi-codegen configuration
- `generate.go` - Go generate directive
- `server.gen.go` - Generated types and MCP server interface (generated)
- `main.go` - Example implementation of the StrictServerInterface

## Generating Code

To generate the MCP server code:

```bash
go generate
```

This will create `server.gen.go` with:
- Type definitions for request/response models
- `StrictServerInterface` with methods for each operation (compatible with strict-server mode)
- `RequestObject` and `ResponseObject` types for each operation
- Parameter types for each operation (path, query, header, cookie, body)
- `RegisterMCPServer` function that works directly with `*mcp.Server`

## Running the Server

### With stdio transport (for MCP clients)

```bash
go run . -transport stdio
```

This is the standard way to run MCP servers. The server communicates over stdin/stdout, which is how MCP clients typically interact with servers.

### With HTTP transport

```bash
go run . -transport http -http-addr :8080
```

This exposes the MCP server over HTTP using the streamable HTTP transport, which can be useful for testing or web-based clients.

## Using with an MCP Client

To use this server with an MCP client like Claude Desktop, you would typically:

1. Build the server:
   ```bash
   go build -o pet-store-mcp
   ```

2. Configure your MCP client to run the server. For Claude Desktop, add to your config:
   ```json
   {
     "mcpServers": {
       "pet-store": {
         "command": "/path/to/pet-store-mcp",
         "args": ["-transport", "stdio"]
       }
     }
   }
   ```

3. The client can now call tools like:
   - "ListPets" with optional query parameters
   - "CreatePet" with a pet object
   - "GetPet" with a pet ID
   - etc.

## Input Schema Structure

The generated MCP tools use a structured input schema that separates different parameter types:

```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "object",
      "properties": { ... }
    },
    "query": {
      "type": "object",
      "properties": { ... }
    },
    "header": {
      "type": "object",
      "properties": { ... }
    },
    "cookie": {
      "type": "object",
      "properties": { ... }
    },
    "body": { ... }
  }
}
```

This structure makes it clear where each parameter comes from in the original HTTP API.

## Implementation

The `main.go` file shows a simple in-memory implementation of the `StrictServerInterface`. In a real application, you would:

1. Implement the `StrictServerInterface` with your business logic
2. Connect to databases, external services, etc.
3. Handle errors appropriately
4. Add authentication/authorization if needed (note: MCP has its own auth mechanisms)

### Key Features

- **No Adapter Needed**: The generated `RegisterMCPServer` function works directly with `*mcp.Server` from the SDK
- **Strict-Server Compatible**: The generated interface uses the same `RequestObject`/`ResponseObject` pattern as strict-server mode
- **Direct Integration**: Simply pass your `*mcp.Server` and your `StrictServerInterface` implementation to `RegisterMCPServer`

```go
server := mcp.NewServer(...)
impl := &MyImplementation{}
RegisterMCPServer(server, impl)
```

## Dependencies

This example requires:

```bash
go get github.com/modelcontextprotocol/go-sdk/mcp
```

The MCP SDK is maintained separately and provides the core MCP protocol implementation.

