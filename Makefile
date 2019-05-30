all: generate_templates install

generate_templates:
	@echo "Generating templates ..."
	@go generate github.com/deepmap/oapi-codegen/pkg/codegen/templates

install:
	@echo "Installing oapi-codegen binary ..."
	@go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen

build:
	@echo "Building oapi-codegen binary ..."
	@go build -o cmd/oapi-codegen/oapi-codegen github.com/deepmap/oapi-codegen/cmd/oapi-codegen

test:
	@echo "Running unit tests..."
	@go test -cover -v ./...
