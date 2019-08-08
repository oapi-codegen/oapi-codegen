#
# Makefile to build the Blue Secrets Service (Certificate Manager)
#
.PHONY: all lint test
.PHONY: debug-build debug
.PHONY: semver clean deps
.PHONY: int-test

#
# Constants and runtime variables
#
PACKAGE				= oapi-codegen
PACKAGE_PATH		= gitswarm.f5net.com/indigo/product/controller-thirdparty
FULL_PACKAGE        = $(PACKAGE_PATH)/$(PACKAGE)
GO_PKGS             = $(shell go list ./... | grep -v test)
GO_PKGS_PATH        = $(shell go list -f '{{.Dir}}' ./... | grep -v test)
DOCKER_TAG       	?= latest
OUT_DIR           	?= build
DEBUG_PACKAGE 		= $(PACKAGE)-debug


MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
INTEGRATION_DIR = "$(MKFILE_DIR)/test/integration"

TEST_TOOL_NAME = "test_tool.sh"
TEST_TOOL = "$(INTEGRATION_DIR)/$(TEST_TOOL_NAME)"

#
# Targets used for local development
#
all: test lint

export OUT_DIR

clean:
	rm -rf ./build_version coverage* test-coverage.html ./deploy test-results.out \
		unit_tests.xml full-lint-report.xml $(OUT_DIR)

#
# Generate the template inline files and *.gen.go files
#
generate:
	go generate ./pkg/...
	go generate ./...
	goimports -w ./

#
# Run all tests and aggregate results into single coverage output.
# TODO: want to include -race flag with go test.
#
ifdef VERBOSE
test: generate
	@echo "mode: set" > coverage-all.out

	@go test -v -timeout 30s -tags unit -coverprofile=coverage.out -race ./... | \
		tee -a test-results.out || exit 1 \
		tail -n +2 coverage.out >> coverage-all.out || exit 1
	@go tool cover -func=coverage-all.out && \
		go tool cover -html=coverage-all.out -o test-coverage.html
else
test: generate
	@echo "mode: set" > coverage-all.out

	@go test -timeout 30s -tags unit -coverprofile=coverage.out -race ./... | \
		tee -a test-results.out || exit 1;\
		tail -n +2 coverage.out >> coverage-all.out || exit 1
	@go tool cover -html=coverage-all.out -o test-coverage.html
endif

#
# Utilities targets
#
# run:
# 	$(OUT_DIR)/$(PACKAGE)
#
# Lint go code non-critical checks
#
lint:
	echo "ignoring lint on this project"

