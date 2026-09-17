GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

# Generated code embeds the OpenAPI spec deflate-compressed, so its bytes
# depend on the toolchain's compress/flate — whose output changed in Go 1.27.
# Every toolchain from the `go` directive in go.mod through 1.26.x produces the
# same blob, so `generate` normally uses whatever Go is installed and leaves a
# GOTOOLCHAIN the environment already sets alone (it is often pinned to `local`
# by policy). Only a toolchain at or past the change is stepped down, to the
# version go.mod declares, so that regenerating does not rewrite the base64 in
# every generated file.
# See https://github.com/oapi-codegen/oapi-codegen/issues/2556
#
# TODO: drop GO_FLATE_CHANGE_MINOR, GO_TOOLCHAIN_FALLBACK and the conditional
# export below once the `go` directive reaches 1.27. Every supported toolchain
# agrees again at that point, and the local one can always be used.
GO_FLATE_CHANGE_MINOR := 27
GO_DIRECTIVE := $(shell awk '/^go /{print $$2; exit}' go.mod)
# GOTOOLCHAIN wants a toolchain version (1.2.3); the `go` directive may be a
# language version (1.2), which is not a valid toolchain name.
GO_TOOLCHAIN := go$(if $(word 3,$(subst ., ,$(GO_DIRECTIVE))),$(GO_DIRECTIVE),$(GO_DIRECTIVE).0)
# Empty unless the Go that would run is at or past the flate change.
GO_TOOLCHAIN_FALLBACK := $(shell go env GOVERSION 2>/dev/null | sed 's/^go//' | \
	awk -F. '($$1>1)||($$1==1 && $$2>=$(GO_FLATE_CHANGE_MINOR)){print "$(GO_TOOLCHAIN)"}')

help:
	@echo "This is a helper makefile for oapi-codegen"
	@echo "Targets:"
	@echo "    generate:    regenerate all generated files"
	@echo "    test:        run all tests"
	@echo "    tidy         tidy go mod"
	@echo "    lint         lint the project"

$(GOBIN)/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/main/install.sh | sh -s -- -b $(GOBIN) v2.13.2

.PHONY: tools
tools: $(GOBIN)/golangci-lint

lint: tools
	# run the root module explicitly, to prevent recursive calls by re-invoking `make ...` top-level
	$(GOBIN)/golangci-lint run ./...
	# then, for all child modules, use a module-managed `Makefile`
	git ls-files '**/*go.mod' -z | xargs -0 -I{} bash -xc 'cd $$(dirname {}) && env GOBIN=$(GOBIN) make lint'

lint-ci: tools
	# for the root module, explicitly run the step, to prevent recursive calls
	$(GOBIN)/golangci-lint run ./... --output.text.path=stdout --timeout=5m
	# then, for all child modules, use a module-managed `Makefile`
	git ls-files '**/*go.mod' -z | xargs -0 -I{} bash -xc 'cd $$(dirname {}) && env GOBIN=$(GOBIN) make lint-ci'

ifneq ($(GO_TOOLCHAIN_FALLBACK),)
generate: export GOTOOLCHAIN := $(GO_TOOLCHAIN_FALLBACK)
endif
generate:
	# for the root module, explicitly run the step, to prevent recursive calls
	go generate ./...
	# then, for all child modules, use a module-managed `Makefile`
	git ls-files '**/*go.mod' -z | xargs -0 -I{} bash -xc 'cd $$(dirname {}) && make generate'

test:
	# for the root module, explicitly run the step, to prevent recursive calls
	go test -cover ./...
	# then, for all child modules, use a module-managed `Makefile`
	git ls-files '**/*go.mod' -z | xargs -0 -I{} bash -xc 'cd $$(dirname {}) && make test'

tidy:
	# for the root module, explicitly run the step, to prevent recursive calls
	go mod tidy
	# then, for all child modules, use a module-managed `Makefile`
	git ls-files '**/*go.mod' -z | xargs -0 -I{} bash -xc 'cd $$(dirname {}) && make tidy'

$(GOBIN)/mdtoc:
	env GOBIN=$(GOBIN) go install sigs.k8s.io/mdtoc@v1.4.0

# Generate/update the Table of Contents in Markdown files.
# Files must contain <!-- toc --> / <!-- /toc --> sentinel comments.
# Use "make readme-toc-check" in CI to verify TOCs are up-to-date.
readme-toc: $(GOBIN)/mdtoc
	$(GOBIN)/mdtoc --inplace README.md

readme-toc-check: $(GOBIN)/mdtoc
	$(GOBIN)/mdtoc --inplace --dryrun README.md

tidy-ci:
	# for the root module, explicitly run the step, to prevent recursive calls
	go mod tidy -diff
	# then, for all child modules, use a module-managed `Makefile`
	git ls-files '**/*go.mod' -z | xargs -0 -I{} bash -xc 'cd $$(dirname {}) && make tidy-ci'
