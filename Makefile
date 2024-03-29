
# Image URL to use all building/pushing image targets
IMG ?= kjournal/apiserver:latest

TEST_PROFILE=elasticsearchv7-fluentbit_elasticsearch-kjournal-structured

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

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

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/apis/..."
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/storage/elasticsearch/..."

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
#test: generate fmt vet envtest ## Run tests
test: ## Run tests
	go test ./... -v -coverprofile coverage.out

.PHONY: configs
configs:
	rm -rf chart/kjournal/config-templates
	cp -Rpv config/base/components/config-templates chart/kjournal/
	find chart/kjournal/config-templates -name kustomization.yaml -exec rm -rfv {} \;
	make configs -C cli

.PHONY: run
run: generate fmt vet ## Run apiserver locally
	go run cmd/*

#-----------------------------------------------------------------------------------------------
##@ Build artifacts

.PHONY: build
build: generate ## Build apiserver binary
	CGO_ENABLED=0 go build -o bin/apiserver cmd/*

.PHONY: docker-build
docker-build: build ## Build docker image with the apiserver
	cp bin/apiserver kjournal-apiserver
	docker build -f Dockerfile.release -t ${IMG} .
	rm kjournal-apiserver

#-----------------------------------------------------------------------------------------------
##@ Documentation

api-docs: gen-crd-api-reference-docs  ## Generate API reference documentation
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir=./pkg/apis/core/v1alpha1 -config=./hack/api-docs/config.json -template-dir=./hack/api-docs/template -out-file=./docs/api/core.kjournal.v1alpha1.md
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir=./pkg/apis/config/v1alpha1 -config=./hack/api-docs/config.json -template-dir=./hack/api-docs/template -out-file=./docs/api/config.kjournal.v1alpha1.md

.PHONY: apiserver-cmdref
apiserver-cmdref: build  ## Build apiserver command line reference
	./bin/apiserver cmdref -d docs/server/cmdref


.PHONY: gen-helm-docs
gen-helm-docs: helm-docs ## Build helm chart readme using helm-docs
	$(HELM_DOCS) -c chart/
	sed 's/# kjournal/Helm Chart/g' -i docs/server/methods/helm.md

.PHONY: docs
docs: api-docs apiserver-cmdref gen-helm-docs mkdocs  ## Build docs using mkdocs
	cp README.md docs/index.md
	cp CONTRIBUTING.md docs/contributing.md
	cp chart/kjournal/README.md docs/server/methods/helm.md
	mkdocs build

.PHONY: mkdocs-serve
mkdocs-serve: mkdocs ## Serve docs locally using mkdocs
	mkdocs serve

#-----------------------------------------------------------------------------------------------
##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: kind-deploy-deps
kind-deploy-deps: ## Deploy dependencies to kind
	kind load docker-image ${IMG} --name kjournal
	make kind-load -C cli
	kustomize build config/tests/${TEST_PROFILE}/dependencies --enable-helm | kubectl apply -f -	

.PHONY: kind-deploy-apiserver
kind-deploy-apiserver: ## Deploy apiserver to kind
	kustomize build config/tests/${TEST_PROFILE}/kjournal | kubectl apply -f -	

.PHONY: kind-deploy-validate
kind-deploy-validate: ## Validate kjournal
	kustomize build config/tests/${TEST_PROFILE}/validation | kubectl apply -f -	
	kubectl -n kjournal-system wait --for=condition=complete --timeout=300s job/validation

.PHONY: kind-debug
kind-debug: docker-build kind-deploy-apiserver kind-deploy-deps ## Deploy to kind and tail log
	kubectl -n kjournal-system rollout restart deployment/kjournal-apiserver
	kubectl -n kjournal-system rollout status deployment/kjournal-apiserver
	kubectl -n kjournal-system logs -l api=kjournal -f

.PHONY: kind-dev-tools
kind-dev-tools: kind-dev-tools ## Deploy dev-tools to kind
	kustomize config/dev-tools | kubectl apply -f -

.PHONY: deploy
deploy: kustomize ## Deploy apiserver to the K8s cluster specified in ~/.kube/config
	cd config/default && $(KUSTOMIZE) edit set image apiserver=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy apiserver from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

#-----------------------------------------------------------------------------------------------
##@ Build dependencies

CONTROLLER_GEN = controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# Find or download gen-crd-api-reference-docs
GEN_CRD_API_REFERENCE_DOCS = $(GOBIN)/gen-crd-api-reference-docs
.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: ## Download gen-crd-api-reference-docs locally if necessary
	$(call go-install-tool,$(GEN_CRD_API_REFERENCE_DOCS),github.com/ahmetb/gen-crd-api-reference-docs@3f29e6853552dcf08a8e846b1225f275ed0f3e3b)

HELM_DOCS = $(GOBIN)/helm-docs
.PHONY: helm-docs
helm-docs: ## Download helm-docs locally if necessary
	$(call go-install-tool,$(HELM_DOCS),github.com/norwoodj/helm-docs/cmd/helm-docs@v1.11.0)

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
