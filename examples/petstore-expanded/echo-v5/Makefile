# This module requires Go 1.25+ (for echo/v5). Skip gracefully on older versions.
GO_VERSION := $(shell go version | sed 's/.*go\([0-9]*\)\.\([0-9]*\).*/\1\2/')
GO_125_OR_LATER := $(shell [ "$(GO_VERSION)" -ge 125 ] 2>/dev/null && echo yes || echo no)

ifeq ($(GO_125_OR_LATER),yes)

lint:
	$(GOBIN)/golangci-lint run ./...

lint-ci:
	$(GOBIN)/golangci-lint run ./... --output.text.path=stdout --timeout=5m

generate:
	go generate ./...

test:
	go test -cover ./...

tidy:
	go mod tidy

tidy-ci:
	tidied -verbose

else

lint generate test tidy lint-ci tidy-ci:
	@echo "Skipping echo-v5 example: requires Go 1.25+ (found $(shell go version))"

endif
