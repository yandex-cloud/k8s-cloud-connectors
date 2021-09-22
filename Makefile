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

CHART_NAME := yandex-cloud-connectors

manifest: ensure-controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects via controller-gen tool.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=connector-manager-role webhook paths="./..." \
				output:crd:artifacts:config=./helm/$(CHART_NAME)/crds \
    			output:rbac:artifacts:config=./helm/$(CHART_NAME)/templates/system \
    			output:webhook:artifacts:config=./helm/$(CHART_NAME)/templates/webhook

generate: ensure-controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="LICENSE" paths="./..."

helm-license: ## Copy project license into helm chart.
	cp ./LICENSE ./helm/$(CHART_NAME)

prepare-chart: manifest helm-license ## Creates all the files needed inside the helm chart.

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

lint: ensure-linter ## Run golangci-lint (https://golangci-lint.run/) against code.
	$(GOLANGCI-LINT) run ./... --verbose

test: generate fmt vet lint ## Run tests for this connector and common packages.
	go test ./... -coverprofile cover.out

##@ Build

local-build-manager: test ## Build manager binary locally.
	go build -o ./bin/manager ./cmd/yc-connector-manager/main.go

local-build-certifier: test ## Build manager binary locally.
	go build -o ./bin/certifier ./cmd/yc-connector-certifier/main.go

local-build: local-build-manager local-build-certifier ## Build all binaries locally.

## Resulting tag of a manager binary
MANAGER_IMG := $(REGISTRY)/manager:$(VERSION)
## Resulting tag of a certifier binary
CERTIFIER_IMG := $(REGISTRY)/certifier:$(VERSION)

docker-check-vars: # Check that all required variables are set
ifndef REGISTRY
	$(error "You must set REGISTRY in order to build docker images")
endif
ifndef VERSION
	$(error "You must set VERSION in order to build docker images")
endif

docker-build-manager: test docker-check-vars ## Build docker image with the manager.
	docker build -t $(MANAGER_IMG) --file manager.dockerfile .

docker-build-certifier: test docker-check-vars ## Build docker image with the certifier.
	docker build -t $(CERTIFIER_IMG) --file certifier.dockerfile .

docker-push-manager: docker-build-manager ## Push docker image with the manager. <REGISTRY> must be specified.
	docker push $(MANAGER_IMG)

docker-push-certifier: docker-build-certifier ## Push docker image with the certifier. <REGISTRY> must be specified.
	docker push $(CERTIFIER_IMG)

docker-push: docker-push-manager docker-push-certifier ## Push all images to docker. <REGISTRY> must be specified.

##@ Deployment

install: prepare-chart ## Deploy to the k8s cluster
	helm install yandex-cloud-connectors helm/yandex-cloud-connectors

uninstall: ## Undeploy from the k8s cluster
	helm uninstall yandex-cloud-connectors

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
