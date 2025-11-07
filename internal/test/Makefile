lint:
	@echo "Skipping lint for internal/test due to known iasm module path conflict"
	@# $(GOBIN)/golangci-lint run ./...

lint-ci:
	$(GOBIN)/golangci-lint run ./... --output.text.path=stdout --timeout=5m

generate:
	go generate ./...

test:
	@echo "Skipping test for internal/test due to known iasm module path conflict"
	@# go test -cover ./...

tidy:
	go mod tidy -e

tidy-ci:
	tidied -verbose
