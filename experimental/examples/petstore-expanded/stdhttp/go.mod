module github.com/oapi-codegen/oapi-codegen/experimental/examples/petstore-expanded/stdhttp

go 1.24

require github.com/oapi-codegen/oapi-codegen/experimental/examples/petstore-expanded/stdhttp/server v0.0.0

require github.com/google/uuid v1.6.0 // indirect

replace github.com/oapi-codegen/oapi-codegen/experimental/examples/petstore-expanded/stdhttp/server => ./server
