lint:
	$(GOBIN)/golangci-lint run ./...

lint-ci:
	$(GOBIN)/golangci-lint run ./... --output.text.path=stdout --timeout=5m

generate:
	go generate ./...

test:
	@echo "Skipping tests in experimental module"

tidy:
	go mod tidy

tidy-ci:
	tidied -verbose
