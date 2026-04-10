SHELL := /bin/bash

.PHONY: default clean build golangci-lintfmt lint shellcheck abcgo openapi-check style run test \
	cover integration_tests license before_commit help godoc install_docgo install_addlicense \
	container-buid container-run

SOURCES:=$(shell find . -name '*.go')
BINARY:=insights-operator-gathering-conditions-service
DOCFILES:=$(addprefix docs/packages/, $(addsuffix .html, $(basename ${SOURCES})))

default: build

clean: ## Run go clean
	@go clean

build: ${BINARY} ## Build binary containing service executable

${BINARY}: ${SOURCES}
	./build.sh

golangci-lint: ## Run golangci-lint
	pre-commit run --all-files golangci-lint-full

fmt: ## Run go formatting
	@echo "Running go formatting"
	golangci-lint fmt

lint: ## Run go liting
	@echo "Running go linting"
	golangci-lint run --fix

shellcheck: ## Run shellcheck
	pre-commit run --all-files shellcheck

abcgo: ## Run ABC metrics checker
	@echo "Run ABC metrics checker"
	pre-commit run --all-files abcgo

openapi-check:  ## Validate the OpenAPI specification files
	./check_openapi.sh

conditions: get_conditions.sh ## Clone the conditions repo and build it to gather the conditions
	./get_conditions.sh

check-config: ${BINARY} conditions ## Check all the configuration files are parsable
	./${BINARY} --check-config

style: golangci-lint abcgo shellcheck check-config ## Run all the formatting related commands (fmt, lint, abc) + check shell scripts

run: ${BINARY} ## Build the project and executes the binary
	./$^

test: ${BINARY} ## Run the unit tests
	./unit-tests.sh

cover: test
	@go tool cover -html=coverage.out

coverage:
	@go tool cover -func=coverage.out

integration_tests: ${BINARY} ## Run all integration tests
	@echo "Running all integration tests"
	@./test.sh

license: install_addlicense
	addlicense -c "Red Hat, Inc" -l "apache" -v ./

before_commit: test integration_tests openapi-check license ## Checks done before commit
	pre-commit run --all-files
	./check_coverage.sh

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

function_list: ${BINARY} ## List all functions in generated binary file
	go tool objdump ${BINARY} | grep ^TEXT | sed "s/^TEXT\s//g"

docs/packages/%.html: %.go
	mkdir -p $(dir $@)
	docgo -outdir $(dir $@) $^
	addlicense -c "Red Hat, Inc" -l "apache" -v $@

godoc: export GO111MODULE=off
godoc: install_docgo install_addlicense ${DOCFILES}

install_docgo: export GO111MODULE=off
install_docgo:
	[[ `command -v docgo` ]] || go get -u github.com/dhconnelly/docgo

install_addlicense: export GO111MODULE=off
install_addlicense:
	[[ `command -v addlicense` ]] || GO111MODULE=off go get -u github.com/google/addlicense


## --------------------------------------
## Go Module
## --------------------------------------

## --------------------------------------
## Container
## --------------------------------------

container-build: ## Build the container image
	$(CONTAINER_RUNTIME) build -t $(CONTAINER_IMAGE_NAME) .

container-run: ## Run the container image
	$(CONTAINER_RUNTIME) run \
		--rm \
		-p 8081:8081 \
		$(CONTAINER_IMAGE_NAME)
