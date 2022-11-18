# Image URL to use all building/pushing image targets
IMG ?= kjournal/cli:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: test build

.PHONY: all
all: build

#-----------------------------------------------------------------------------------------------
##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

#-----------------------------------------------------------------------------------------------
##@ Development

.PHONY: tidy
tidy:  ## Run go mod tidy against code
	go mod tidy

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint against code
	$(GOLANGCI_LINT) run --timeout=2m ./...

.PHONY: test
test: ## Run tests
	go test ./... -v -coverprofile coverage.out

.PHONY: configs
configs:
	rm -rfv cmd/config
	mkdir cmd/config
	cp -Rpv ../config/base cmd/config/

.PHONY: run
run: generate fmt vet ## Run cli
	go run cmd/*

#-----------------------------------------------------------------------------------------------
##@ Build artifacts

build: configs fmt vet lint
	CGO_ENABLED=0 go build -o ./bin/kjournal ./cmd

.PHONY: docker-build
docker-build: build ## Build docker image with the cli
	cp bin/kjournal kjournal-cli
	docker build -f Dockerfile.release -t ${IMG} .
	rm kjournal-cli

#-----------------------------------------------------------------------------------------------
##@ Install

.PHONY: kind-load
kind-load: ## Load image into kind.
	kind load docker-image ${IMG} --name kjournal

.PHONY: install
install:
	CGO_ENABLED=0 go install ./cmd

#-----------------------------------------------------------------------------------------------
##@ Build dependencies

GOLANGCI_LINT = $(GOBIN)/golangci-lint
.PHONY: golangci-lint
golangci-lint: ## Download golint locally if necessary
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0)

.PHONY: mkdocs
mkdocs: ## Install mkdocs
	pip install mkdocs
	pip install mkdocs-minify-plugin
	pip install mkdocs-material

# go-install-tool will 'go install' any package $2 and install it to $1
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
env -i bash -c "GOBIN=$(GOBIN) PATH=$(PATH) GOPATH=$(shell go env GOPATH) GOCACHE=$(shell go env GOCACHE) go install $(2)" ;\
rm -rf $$TMP_DIR ;\
}
endef
