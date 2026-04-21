GO_VERSION := $(shell go version | sed -E 's/.*go([0-9]+\.[0-9]+).*/\1/')
MIN_VERSION := 1.25
VERSION_OK := $(shell printf '%s\n%s' '$(MIN_VERSION)' '$(GO_VERSION)' | sort -V | head -1 | grep -q '^$(MIN_VERSION)$$' && echo yes || echo no)

lint:
ifeq ($(VERSION_OK),yes)
	$(GOBIN)/golangci-lint run ./...
else
	@echo "Skipping lint: requires Go >= $(MIN_VERSION), have $(GO_VERSION)"
endif

lint-ci:
ifeq ($(VERSION_OK),yes)
	$(GOBIN)/golangci-lint run ./... --output.text.path=stdout --timeout=5m
else
	@echo "Skipping lint-ci: requires Go >= $(MIN_VERSION), have $(GO_VERSION)"
endif

generate:
ifeq ($(VERSION_OK),yes)
	go generate ./...
else
	@echo "Skipping generate: requires Go >= $(MIN_VERSION), have $(GO_VERSION)"
endif

test:
ifeq ($(VERSION_OK),yes)
	go test -cover ./...
else
	@echo "Skipping test: requires Go >= $(MIN_VERSION), have $(GO_VERSION)"
endif

tidy:
ifeq ($(VERSION_OK),yes)
	go mod tidy
else
	@echo "Skipping tidy: requires Go >= $(MIN_VERSION), have $(GO_VERSION)"
endif

tidy-ci:
ifeq ($(VERSION_OK),yes)
	tidied -verbose
else
	@echo "Skipping tidy-ci: requires Go >= $(MIN_VERSION), have $(GO_VERSION)"
endif
