SHELL:=/bin/bash

YELLOW := \e[0;33m
RESET := \e[0;0m

GOVER := $(shell go env GOVERSION)
GOMINOR := $(shell bash -c "cut -f2 -d. <<< $(GOVER)")

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
	$(call execute-if-go-124,$(GOBIN)/golangci-lint run ./...)

lint-ci:

	$(call execute-if-go-124,$(GOBIN)/golangci-lint run ./... --out-format=colored-line-number --timeout=5m)

generate:
	$(call execute-if-go-124,go generate ./...)

test:
	$(call execute-if-go-124,go test -cover ./...)

tidy:
	$(call execute-if-go-124,go mod tidy)

tidy-ci:
	$(call execute-if-go-124,tidied -verbose)
