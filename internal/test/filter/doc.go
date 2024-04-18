package client

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --package=filtertags --generate=server -o tags/server.gen.go -include-tags included-tag1,included-tag2 server.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --package=filteroperations --generate=server -o operations/server.gen.go -include-operation-ids included-operation1,included-operation2 server.yaml
