# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Service
SERVICE_NAME = insights-conditional-service
BIN_NAME = $(SERVICE_NAME)

# Container
CONTAINER_NAME = $(SERVICE_NAME)
CONTAINER_NAMESPACE ?= redhatinsights
CONTAINER_TAG = $(shell git describe --tags --exact-match 2>/dev/null || echo latest)
CONTAINER_IMAGE_NAME = ${CONTAINER_NAMESPACE}/${CONTAINER_NAME}:${CONTAINER_TAG}

# Testing
GO_TEST_FLAGS = $(VERBOSE)
COVER_PROFILE = cover.out

# Configuration
RUN_FLAGS ?= 

# Tools
CONTAINER_RUNTIME := $(shell command -v podman 2> /dev/null || echo docker)
GOLANGCI_LINT := $(GOBIN)/golangci-lint

export GO111MODULE=on
export GOFLAGS=-mod=vendor

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: githooks
githooks: ## Configure the repository to use the git hooks
	git config core.hooksPath ./.githooks

## --------------------------------------
## Tests
## --------------------------------------

# Run the tests
.PHONY: test
test: unit ## Run all the tests

# Run the unit tests
.PHONY: unit
unit: ## Run the unit tests
	go test $(GO_TEST_FLAGS) -coverprofile $(COVER_PROFILE) ./...

.PHONY: coverage
coverage:
	./.citools/check-coverage.sh

.PHONE: unit-verbose
unit-verbose:
	VERBOSE=-v make unit

## --------------------------------------
## Linting
## --------------------------------------

.PHONY: precommit
precommit: ## Executes the pre-commit hook (check the stashed changes)
	./.githooks/pre-commit

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Executes the linting tool (vet, sec, and others)
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT) ## Executes the linting with fix
	$(GOLANGCI_LINT) run --fix $(RUN_FLAGS)

$(GOLANGCI_LINT):
	./.citools/install-golangci-lint.sh

## --------------------------------------
## Build/Run
## --------------------------------------

.PHONY: run
run: ## Executes the service
	go run ./cmd/server/main.go $(RUN_FLAGS)

.PHONY: build
build: ## Compiles the service
	go build -o ./bin/$(BIN_NAME) ./cmd/server

## --------------------------------------
## Container
## --------------------------------------

.PHONY: container-buid
container-build: ## Build the container image
	$(CONTAINER_RUNTIME) build -t $(CONTAINER_IMAGE_NAME) .

.PHONY: container-run
container-run: ## Run the container image
	$(CONTAINER_RUNTIME) run \
		--rm \
		--name $(CONTAINER_NAME) \
		-p 8081:8081 \
		$(CONTAINER_IMAGE_NAME)

## --------------------------------------
## Go Module
## --------------------------------------

.PHONY: vendor
vendor: ## Runs tiny, vendor and verify the module
	go mod tidy
	go mod vendor
	go mod verify