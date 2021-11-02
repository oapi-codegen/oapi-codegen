help:
	@echo "This is a helper makefile for oapi-codegen"
	@echo "Targets:"
	@echo "    generate:    regenerate all generated files"
	@echo "    test:        run all tests"
	@echo "    gin_example  generate gin example server code"
	@echo "    tidy         tidy go mod"

generate:
	go generate ./pkg/...
	go generate ./...

test:
	go test -cover ./...

gin_example:
	@echo "generate gin example...."
	go run cmd/oapi-codegen/oapi-codegen.go --config=examples/petstore-expanded/gin/api/server.cfg.yaml examples/petstore-expanded/petstore-expanded.yaml
	go run cmd/oapi-codegen/oapi-codegen.go --config=examples/petstore-expanded/gin/api/types.cfg.yaml examples/petstore-expanded/petstore-expanded.yaml

tidy:
	@echo "tidy..."
	go mode tidy
