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

tidy:
	@echo "tidy..."
	go mod tidy
