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

help: ## Display this message.
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
			output:rbac:artifacts:config=./config/system \
			output:webhook:artifacts:config=./config/webhook

generate: ensure-controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="LICENSE" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

lint: ensure-linter ## Run golangci-lint (https://golangci-lint.run/) against code.
	$(GOLANGCI-LINT) run ./... --verbose

test: manifests generate fmt vet lint ## Run tests for this connector and common packages.
	go test ./... -coverprofile cover.out

##@ Build

local-build-manager: test ## Build manager binary locally.
	go build -o ./bin/manager ./cmd/yc-connector-manager/main.go

local-build-certifier: test ## Build manager binary locally.
	go build -o ./bin/certifier ./cmd/yc-connector-certifier/main.go

local-build: local-build-manager local-build-certifier ## Build all binaries locally.

## Version of an images, can be set up externally
IMG_TAG ?= latest

## Image name of a manager binary
MANAGER_IMG_NAME := yc-connector-manager
## Resulting tag of a manager binary
MANAGER_IMG := $(MANAGER_IMG_NAME):$(IMG_TAG)
docker-build-manager: test ## Build docker image with the manager.
	docker build -t $(MANAGER_IMG) --file manager.dockerfile .

## Image name of a certifier binary
CERTIFIER_IMG_NAME := yc-connector-certifier
## Resulting tag of a certifier binary
CERTIFIER_IMG := $(CERTIFIER_IMG_NAME):$(IMG_TAG)
docker-build-certifier: test ## Build docker image with the certifier.
	docker build -t $(CERTIFIER_IMG) --file certifier.dockerfile .

docker-push-manager: docker-build-manager ## Push docker image with the manager. <REGISTRY> must be specified.
ifndef REGISTRY
	$(error "You must set REGISTRY in order to push")
endif
	docker tag $(MANAGER_IMG) $(REGISTRY)/$(MANAGER_IMG)
	docker push $(REGISTRY)/$(MANAGER_IMG)

docker-push-certifier: docker-build-certifier ## Push docker image with the certifier. <REGISTRY> must be specified.
ifndef REGISTRY
	$(error "You must set REGISTRY in order to push")
endif
	docker tag $(CERTIFIER_IMG) $(REGISTRY)/$(CERTIFIER_IMG)
	docker push $(REGISTRY)/$(CERTIFIER_IMG)

docker-push: docker-push-manager docker-push-certifier ## Push all images to docker. <REGISTRY> must be specified.

##@ Deployment

install: manifests ## Deploy to the k8s cluster specified in ~/.kube/config.
	kubectl apply -k ./config/base

uninstall: ## Undeploy from the k8s cluster specified in ~/.kube/config.
	kubectl delete -k ./config/base

##@ Dependencies

GOLANGCI-LINT := $(ROOT)/bin/golangci-lint ## Location of golangci-lint binary
ensure-linter: ## Download golangci-lint if necessary.
	@if [ ! -x "$(command -v golangci-lint)" ] && [ ! -x $(GOLANGCI-LINT) ]; \
 	then \
  		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
                         | sh -s -- -b $(ROOT)/bin v1.41.0; \
  	fi

CONTROLLER_GEN := $(ROOT)/bin/controller-gen ## Location of controller-gen binary
ensure-controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.0)

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
