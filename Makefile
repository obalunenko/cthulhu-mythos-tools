BIN_DIR=./bin

SHELL := env VERSION=$(VERSION) $(SHELL)
VERSION ?= $(shell git describe --tags $(git rev-list --tags --max-count=1))

APP_NAME?=cthulhu-mythos-tools
SHELL := env APP_NAME=$(APP_NAME) $(SHELL)

RELEASE_BRANCH?=master
SHELL := env RELEASE_BRANCH=$(RELEASE_BRANCH) $(SHELL)

COMPOSE_TOOLS_FILE=deployments/docker-compose/go-tools-docker-compose.yml
COMPOSE_TOOLS_CMD_BASE=docker compose -f $(COMPOSE_TOOLS_FILE)
COMPOSE_TOOLS_CMD_UP=$(COMPOSE_TOOLS_CMD_BASE) up --remove-orphans --exit-code-from
COMPOSE_TOOLS_CMD_PULL=$(COMPOSE_TOOLS_CMD_BASE) build

GOVERSION:=1.24

CTHULHU_MYTHOS_TOOLS_IMAGE?=$(APP_NAME)
SHELL := env CTHULHU_MYTHOS_TOOLS_IMAGE=$(CTHULHU_MYTHOS_TOOLS_IMAGE) $(SHELL)

GOTOOLS_IMAGE?=go-tools-local
SHELL := env GOTOOLS_IMAGE=$(GOTOOLS_IMAGE) $(SHELL)

DOCKER_REPO?=
SHELL := env DOCKER_REPO=$(DOCKER_REPO) $(SHELL)

TARGET_MAX_CHAR_NUM=20

## Show help
help:
	${call colored, help is running...}
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  %-$(TARGET_MAX_CHAR_NUM)s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## Build project.
build: sync-vendor generate compile-app
.PHONY: build

## Compile app.
compile-app:
	./scripts/build/app.sh
.PHONY: compile-app

## Test coverage report.
test-cover:
	./scripts/tests/coverage.sh
.PHONY: test-cover

prepare-cover-report: test-cover
	$(COMPOSE_TOOLS_CMD_UP) prepare-cover-report prepare-cover-report
.PHONY: prepare-cover-report

## Open coverage report.
open-cover-report: prepare-cover-report
	./scripts/open-coverage-report.sh
.PHONY: open-cover-report

## Update readme coverage.
update-readme-cover: build prepare-cover-report
	$(COMPOSE_TOOLS_CMD_UP) update-readme-coverage update-readme-coverage
.PHONY: update-readme-cover

## Run tests.
test:
	$(COMPOSE_TOOLS_CMD_UP) run-tests run-tests
.PHONY: test

test-integration:
	./scripts/tests/integration.sh
.PHONY: test-integration

## Run regression tests.
test-regression: test
.PHONY: test-regression

## Sync vendor and install needed tools.
configure: sync-vendor install-tools

## Sync vendor with go.mod.
sync-vendor:
	./scripts/sync-vendor.sh
.PHONY: sync-vendor

## Fix imports sorting.
imports:
	$(COMPOSE_TOOLS_CMD_UP) fix-imports fix-imports
.PHONY: imports

## Format code with go fmt.
fmt:
	$(COMPOSE_TOOLS_CMD_UP) fix-fmt fix-fmt
.PHONY: fmt

## Format code and sort imports.
format-project: fmt imports
.PHONY: format-project

## Installs vendored tools.
install-tools:
	./scripts/install/vendored-tools.sh
.PHONY: install-tools

## vet project
vet:
	$(COMPOSE_TOOLS_CMD_UP) vet vet
.PHONY: vet

## Run full linting
lint-full:
	$(COMPOSE_TOOLS_CMD_UP) lint-full lint-full
.PHONY: lint-full

## Run linting for build pipeline
lint-pipeline:
	$(COMPOSE_TOOLS_CMD_UP) lint-pipeline lint-pipeline
.PHONY: lint-pipeline

## Run linting for sonar report
lint-sonar:
	$(COMPOSE_TOOLS_CMD_UP) lint-sonar lint-sonar
.PHONY: lint-sonar

## recreate all generated code and documentation.
codegen:
	$(COMPOSE_TOOLS_CMD_UP) go-generate go-generate
.PHONY: codegen

## recreate all generated code and swagger documentation and format code.
generate: codegen format-project vet
.PHONY: generate

## Release
release:
	$(COMPOSE_TOOLS_CMD_UP) release release
.PHONY: release

## Release local snapshot
release-local-snapshot:
	$(COMPOSE_TOOLS_CMD_UP) release-local-snapshot release-local-snapshot
.PHONY: release-local-snapshot

## Check goreleaser config.
check-releaser:
	$(COMPOSE_TOOLS_CMD_UP) release-check-config release-check-config
.PHONY: check-releaser

## Issue new release.
new-version: vet test-regression build
	./scripts/release/new-version.sh
.PHONY: new-release

bump-go-version:
	./scripts/bump-go.sh $(GOVERSION)
.PHONY: bump-go-version

build-go-tools: install-tools
.PHONY: build-go-tools

build-cthulhu-mythos-tools:
	./scripts/build/docker.sh
.PHONY: build-cthulhu-mythos-tools

run-local: build-cthulhu-mythos-tools
	docker compose -f ./deployments/docker-compose/cthulhu-compose.yaml up --build --remove-orphans --detach
.PHONY: run-local

stop-local:
	docker compose -f ./deployments/docker-compose/cthulhu-compose.yaml down
.PHONY: stop-local

down-local:
	docker compose -f ./deployments/docker-compose/cthulhu-compose.yaml down --volumes
.PHONY: down-local

.DEFAULT_GOAL := help

