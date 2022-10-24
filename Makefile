
# Image URL to use all building/pushing image targets
IMG ?= kjournal/apiserver:latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

BINARY_NAME=mybinary

KIND_TEST_PROFILE=elasticsearchv7-fluentbit-kjournal-structured

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

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
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/apis/..."
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/storage/elasticsearch/..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: golangci-lint
golint: ## Download golint locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0)

.PHONY: lint
lint: golangci-lint ## Run golangci-lint against code.
	golangci-lint run ./...

.PHONY: test
#test: generate fmt vet envtest ## Run tests.
test:
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -v -coverprofile coverage.out

.PHONY: kind-deploy
kind-deploy: docker-build ## Deploy to kind.
	kind load docker-image ${IMG} --name kjournal
	kustomize build config/tests/${KIND_TEST_PROFILE} --enable-helm | kubectl apply -f -
	kubectl -n kjournal-system wait deployments --all --for=condition=available --timeout=120s
	kubectl -n kjournal-system get pods | grep -v 'test\|NAME\|fluent\|log-generator' | cut -d ' ' -f1 | xargs kubectl wait pods -n kjournal-system --all --for=condition=ready  --timeout=120s

.PHONY: kind-debug
kind-debug: kind-deploy ## Deploy to kind and tail log
	kubectl -n kjournal-system rollout restart deployment/kjournal-apiserver
	kubectl -n kjournal-system rollout status deployment/kjournal-apiserver
	kubectl -n kjournal-system logs -l api=kjournal -f

.PHONY: kind-dev-tools
kind-dev-tools: kind-dev-tools ## Deploy dev-tools to kind.
	kustomize config/dev-tools | kubectl apply -f -

##@ Build

.PHONY: prepare-embeds
prepare-embed:
	rm -rf cli/cmd/config
	mkdir -p cli/cmd/config
	cp -rpv config/{apiserver,components,prometheus,namespace,rbac} cli/cmd/config/

.PHONY: build
build: generate fmt vet ## Build apiserver binary.
	CGO_ENABLED=0 go build -o bin/apiserver cmd/*

.PHONY: run
run: generate fmt vet ## Run apiserver from your host.
	go run cmd/*

.PHONY: docker-build
docker-build: build ## Build docker image with the apiserver.
	cp bin/apiserver kjournal-apiserver
	docker build -f Dockerfile.release -t ${IMG} .
	rm kjournal-apiserver

.PHONY: docker-push
docker-push: ## Push docker image with the apiserver.
	docker push ${IMG}

api-docs: gen-crd-api-reference-docs  ## Generate API reference documentation
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir=./pkg/apis/core/v1alpha1 -config=./hack/api-docs/config.json -template-dir=./hack/api-docs/template -out-file=./docs/api/core.kjournal.v1alpha1.md
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir=./pkg/apis/config/v1alpha1 -config=./hack/api-docs/config.json -template-dir=./hack/api-docs/template -out-file=./docs/api/config.kjournal.v1alpha1.md

# Find or download gen-crd-api-reference-docs
GEN_CRD_API_REFERENCE_DOCS = $(GOBIN)/gen-crd-api-reference-docs
.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: ## Download gen-crd-api-reference-docs locally if necessary
	$(call go-install-tool,$(GEN_CRD_API_REFERENCE_DOCS),github.com/ahmetb/gen-crd-api-reference-docs@3f29e6853552dcf08a8e846b1225f275ed0f3e3b)

.PHONY: apiserver-cmdref
apiserver-cmdref: build  ## Build apiserver command line reference
	./bin/apiserver cmdref -d docs/server/cmdref

.PHONY: helm-docs
helm-docs:
	helm-docs -c chart/

.PHONY: gen-docs
gen-docs: api-docs apiserver-cmdref helm-docs mkdocs  ## Build docs using mkdocs
	cp README.md docs/index.md
	cp CONTRIBUTING.md docs/contributing.md
	cp chart/kjournal/README.md docs/server/methods/helm.md
	mkdocs build

.PHONY: mkdocs
mkdocs: ## Install mkdocs
	pip install mkdocs
	pip install mkdocs-minify-plugin
	pip install mkdocs-material

.PHONY: mkdocs-serve
mkdocs-serve: mkdocs ## Serve docs locally using mkdocs
	mkdocs serve

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: deploy
deploy: kustomize ## Deploy apiserver to the K8s cluster specified in ~/.kube/config.
	cd config/default && $(KUSTOMIZE) edit set image apiserver=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy apiserver from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

CONTROLLER_GEN = controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)


# go-install-tool will 'go install' any package $2 and install it to $1.
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
