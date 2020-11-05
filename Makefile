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

build_release_assets: generate test
	mkdir -p ./bin
	GOOS=darwin GOARCH=amd64 go build -o ./.bin/oapi-codegen_darwin_amd64 ./cmd/oapi-codegen/oapi-codegen.go
	GOOS=windows GOARCH=amd64 go build -o ./.bin/oapi-codegen_windows_amd64 ./cmd/oapi-codegen/oapi-codegen.go
	GOOS=linux GOARCH=amd64 go build -o ./.bin/oapi-codegen_linux_amd64 ./cmd/oapi-codegen/oapi-codegen.go
	GOOS=freebsd GOARCH=amd64 go build -o ./.bin/oapi-codegen_freebsd_amd64 ./cmd/oapi-codegen/oapi-codegen.go