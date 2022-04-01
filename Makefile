PROJECT_NAME := "token-parser"
VERSION      ?= $$(git describe --tags | cut -d '-' -f 1)
PKG          := "github.com/sjbodzo/$(PROJECT_NAME)"

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: all
all: build

##@ Development

.PHONY: test
test: ## Run tests
	@go test -cover ./...
	@go test ./... -coverprofile=cover.out && go tool cover -html=cover.out -o coverage.html

.PHONY: build ## Builds the binary locally for the major platforms
build: mac windows linux

mac: ## Builds the binary for mac
	go build -a -ldflags "-w -X '$(PKG)/cmd.Version=$(VERSION)'"  -o bin/$(PROJECT_NAME)-darwin

windows: ## Builds the binary for windows
	go build -a -ldflags "-w -X '$(PKG)/cmd.Version=$(VERSION)'"  -o bin/$(PROJECT_NAME).exe

linux: ## Builds the binary for linux
	go build -a -ldflags "-w -X '$(PKG)/cmd.Version=$(VERSION)'"  -o bin/$(PROJECT_NAME)

.PHONY: docker-build
docker-build: ## Builds the docker image locally
	docker build -t ${PKG}:${VERSION} .

.PHONY: docker-push
docker-push: ## Pushes the docker image to the registry
	docker push ${PKG}:${VERSION}

