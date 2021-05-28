# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"
# Project directory
ROOT := $(shell pwd)

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
	@awk 'BEGIN { \
		FS = ":.*##";\
		printf "Usage:\n  make \033[36m<target>\033[0m\n" \
		} \
		/^[a-zA-Z_0-9-]+:.*?##/ \
		{ \
		printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 \
		} \
		/^##@/ \
		{ \
		printf "\n\033[1m%s\033[0m\n", substr($$0, 5) \
		}' $(MAKEFILE_LIST)

##@ Development

manifests: ensure-controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects via controller-gen tool.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=connector-manager-role webhook paths="./..." \
			output:crd:artifacts:config=./config/base/crd \
			output:rbac:artifacts:config=./config/system

generate: ensure-controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="LICENSE" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

lint: ensure-linter ## Run golangci-lint (https://golangci-lint.run/) against code.
	$(GOLANGCI-LINT) run ./... --verbose

test: generate fmt vet lint ## Run tests for this connector and common packages.
	go test ./... -coverprofile cover.out

##@ Build

build: manifests generate test ## Build manager binary.
	go build -o ./bin/manager ./cmd/yc-connector-manager/main.go

## Image name of a manager binary
IMG_NAME := yc-connector-manager
## Version of a manager binary, can be set up externally
IMG_TAG ?= latest
## Resulting tag of a manager binary
IMG := $(IMG_NAME):$(IMG_TAG)
docker-build: build ## test Build docker image with the manager.
	docker build -t $(IMG) .

docker-push: docker-build ## Push docker image with the manager.
ifndef REGISTRY
	$(error "You must set REGISTRY in order to push")
endif
	docker tag $(IMG) $(REGISTRY)/$(IMG)
	docker push $(REGISTRY)/$(IMG)

##@ Deployment

install: manifests ensure-kustomize ## Deploy controller to the k8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build ./config/base | kubectl apply -f -

uninstall: ensure-kustomize ## Undeploy controller from the k8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build ./config/base | kubectl delete -f -

##@ Dependencies

GOLANGCI-LINT := $(ROOT)/bin/golangci-lint ## Location of golangci-lint binary
ensure-linter: ## Download golangci-lint if necessary.
	@if [ ! -x "$(command -v golangci-lint)" ] && [ ! -x $(GOLANGCI-LINT) ]; \
 	then \
  		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
                         | sh -s -- -b $(ROOT)/bin v1.39.0; \
  	fi

CONTROLLER_GEN := $(ROOT)/bin/controller-gen ## Location of controller-gen binary
ensure-controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0)

KUSTOMIZE := $(ROOT)/bin/kustomize ## Location of kustomize binary
ensure-kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.0.5)

define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(ROOT)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef ## go-get-tool will 'go get' any package $2 and install it to $1
