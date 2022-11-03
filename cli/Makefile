# Image URL to use all building/pushing image targets
IMG ?= kjournal/cli:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

rwildcard=$(foreach d,$(wildcard $(addsuffix *,$(1))),$(call rwildcard,$(d)/,$(2)) $(filter $(subst *,%,$(2)),$(d)))

all: test build

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

.PHONY: test
test:
	go test ./... -v -coverprofile coverage.out

.PHONY: golangci-lint
golint: ## Download golint locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0)

.PHONY: lint
lint: golangci-lint ## Run golangci-lint against code.
	golangci-lint run ./...

manifests:
	rm -rfv cmd/config
	mkdir cmd/config
	cp -Rpv ../config/base cmd/config/

build: manifests fmt vet lint
	CGO_ENABLED=0 go build -o ./bin/kjournal ./cmd

.PHONY: docker-build
docker-build: build ## Build docker image with the cli.
	cp bin/kjournal kjournal-cli
	docker build -f Dockerfile.release -t ${IMG} .
	rm kjournal-cli

.PHONY: kind-load
kind-load: ## Load image into kind.
	kind load docker-image ${IMG} --name kjournal

.PHONY: docker-push
docker-push: ## Push docker image with the cli.
	docker push ${IMG}

.PHONY: install
install:
	CGO_ENABLED=0 go install ./cmd

install-dev:
	CGO_ENABLED=0 go build -o /usr/local/bin ./cmd

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
