SHELL := /bin/bash

.PHONY: default clean build fmt lint vet cyclo ineffassign shellcheck errcheck goconst gosec abcgo json-check openapi-check style run test test-postgres cover integration_tests rest_api_tests sqlite_db license before_commit help godoc install_docgo install_addlicense

SOURCES:=$(shell find . -name '*.go')
BINARY:=insights-operator-gathering-conditions-service
DOCFILES:=$(addprefix docs/packages/, $(addsuffix .html, $(basename ${SOURCES})))

default: build

clean: ## Run go clean
	@go clean

build: ${BINARY} ## Build binary containing service executable

${BINARY}: ${SOURCES}
	./build.sh

fmt: ## Run go fmt -w for all sources
	@echo "Running go formatting"
	./gofmt.sh

lint: ## Run golint
	@echo "Running go lint"
	./golint.sh

vet: ## Run go vet. Report likely mistakes in source code
	@echo "Running go vet"
	./govet.sh

cyclo: ## Run gocyclo
	@echo "Running gocyclo"
	./gocyclo.sh

ineffassign: ## Run ineffassign checker
	@echo "Running ineffassign checker"
	./ineffassign.sh

shellcheck: ## Run shellcheck
	./shellcheck.sh

errcheck: ## Run errcheck
	@echo "Running errcheck"
	./goerrcheck.sh

goconst: ## Run goconst checker
	@echo "Running goconst checker"
	./goconst.sh ${VERBOSE}

gosec: ## Run gosec checker
	@echo "Running gosec checker"
	./gosec.sh ${VERBOSE}

abcgo: ## Run ABC metrics checker
	@echo "Run ABC metrics checker"
	./abcgo.sh ${VERBOSE}

openapi-check:  ## Validate the OpenAPI specification files
	./check_openapi.sh

conditions:  ## Clone the conditions repo and build it to gather the conditions
	if [ ! -d 'insights-operator-gathering-conditions' ]; then git clone https://github.com/RedHatInsights/insights-operator-gathering-conditions; fi
	cd insights-operator-gathering-conditions && ./build.sh
	cp -r insights-operator-gathering-conditions/build conditions

check-config: ${BINARY} conditions ## Check all the configuration files are parsable
	./${BINARY} --check-config

style: fmt vet lint cyclo shellcheck errcheck goconst gosec ineffassign abcgo check-config ## Run all the formatting related commands (fmt, vet, lint, cyclo) + check shell scripts

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

before_commit: style test integration_tests openapi-check license ## Checks done before commit
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

.PHONY: container-buid
container-build: ## Build the container image
	$(CONTAINER_RUNTIME) build -t $(CONTAINER_IMAGE_NAME) .

.PHONY: container-run
container-run: ## Run the container image
	$(CONTAINER_RUNTIME) run \
		--rm \
		-p 8081:8081 \
		$(CONTAINER_IMAGE_NAME)
