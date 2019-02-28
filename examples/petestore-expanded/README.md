OpenAPI Code Generation Example
-------------------------------

This directory contains an example server using our code generator which implements
the OpenAPI [petstore-expanded](https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore-expanded.yaml)
example.

This is the structure:
- `api/`: Contains the OpenAPI 3.0 specification
- `api/petstore/`: The generated code for our pet store handlers
- `internal/`: Pet store handler implementation and unit tests
- `cmd/`: Runnable server implementing the OpenAPI 3 spec.

To generate the handler glue, run:

    go run cmd/oapi-codegen/oapi-codegen.go --package petstore examples/oapi-codegen/api/petstore-expanded.yaml  > examples/oapi-codegen/api/petstore.gen.go
