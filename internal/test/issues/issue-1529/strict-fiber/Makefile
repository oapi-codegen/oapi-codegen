SHELL:=/bin/bash

YELLOW := \e[0;33m
RESET := \e[0;0m

GOVER := $(shell go env GOVERSION)
GOMINOR := $(shell bash -c "cut -f1 -d' ' <<< \"$(GOVER)\" | cut -f2 -d.")

define execute-if-go-124
@{ \
if [[ 24 -le $(GOMINOR) ]]; then \
	$1; \
else \
	echo -e "$(YELLOW)Skipping task as you're running Go v1.$(GOMINOR).x which is < Go 1.24, which this module requires$(RESET)"; \
fi \
}
endef

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
