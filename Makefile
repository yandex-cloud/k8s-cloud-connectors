# ==================================================================== #
# Before you run this script be wary to set up your CONNECTOR variable #
# ==================================================================== #

# Image URL to use all building/pushing image targets
IMG ?= "controller-$(CONNECTOR):latest"
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

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

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./connectors/$(CONNECTOR)/..." output:crd:artifacts:config=./connectors/$(CONNECTOR)/config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="LICENSE" paths="./connectors/$(CONNECTOR)/..."

fmt: ## Run go fmt against code.
	go fmt ./connectors/$(CONNECTOR)/...

vet: ## Run go vet against code.
	go vet ./connectors/$(CONNECTOR)/...

lint: ensure-linter ## Run golangci-lint (https://golangci-lint.run/) against code.
	$(GOLANGCI-LINT) run ./connectors/$(CONNECTOR)/...

ENVTEST_ASSETS_DIR=$(shell pwd)/connectors/$(CONNECTOR)/testbin
test: manifests generate fmt vet lint ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./connectors/$(CONNECTOR)/... -coverprofile ./connectors/$(CONNECTOR)/cover.out

##@ Build

build: generate fmt vet lint ## Build manager binary.
	go build -o ./connectors/$(CONNECTOR)/bin/manager ./connectors/$(CONNECTOR)/cmd/$(CONNECTOR)-controller/main.go

run: manifests generate fmt vet lint ## Run a controller from your host.
	go run ./connectors/$(CONNECTOR)/main.go

# TODO (covariance) uncomment test when tests would be built
docker-build: ## test Build docker image with the manager.
	docker build -t ${IMG} . --build-arg connector=$(CONNECTOR)

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build ./connectors/$(CONNECTOR)/config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build ./connectors/$(CONNECTOR)/config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build ./connectors/$(CONNECTOR)/config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build ./connectors/$(CONNECTOR)/config/default | kubectl delete -f -

##@ Dependencies

GOLANGCI-LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI-LINT-DOWNLOAD = $(shell pwd)/bin
ensure-linter: ## Download golangci-lint if necessary.
	if [ ! -x "$(command -v golangci-lint)" ] && [ ! -x $(GOLANGCI-LINT) ]; \
 	then \
  		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
                         | sh -s -- -b $(GOLANGCI-LINT-DOWNLOAD) v1.39.0; \
  	fi

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.0.5)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef