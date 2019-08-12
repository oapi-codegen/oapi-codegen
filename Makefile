help:
	@echo "This is a helper makefile for oapi-codegen"
	@echo "Targets:"
	@echo "    generate:    regenerate all generated files"
	@echo "    test:        run all tests"

generate:
	go generate ./pkg/...
	go generate ./...

test:
	go test -cover ./...
