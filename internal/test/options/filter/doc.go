// Package optionsfilter drives generation of the operations/ and tags/ sub-packages,
// which exercise output filtering by operationId and by tag respectively. The actual
// generated servers and their tests live in those sub-packages (each is its own Go
// package because each emits a colliding ServerInterface).
package optionsfilter

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --package=optionsfiltertags --generate=server -o tags/server.gen.go -include-tags included-tag1,included-tag2 server.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --package=optionsfilteroperations --generate=server -o operations/server.gen.go -include-operation-ids included-operation1,included-operation2 server.yaml
