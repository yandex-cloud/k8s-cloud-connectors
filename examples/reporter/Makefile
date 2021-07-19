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

##@ Build

build-server: ## Builds server container and pushes it to the specified registry
	docker build -t ${REGISTRY}/ycc-example/server:latest --file server.dockerfile .
    docker push ${REGISTRY}/ycc-example/server:latest

build-worker: ## Builds worker container and pushes it to the specified registry
	docker build -t ${REGISTRY}/ycc-example/worker:latest --file worker.dockerfile .
    docker push ${REGISTRY}/ycc-example/worker:latest

build-all: build-server build-worker ## Builds both server and worker

create-sa:
	SAID=$(yc iam service-account create ycc-example-sa --format json | jq -r '.id')
	yc resource-manager folder add-access-binding --id "${FOLDER_ID}" --role ymq.admin --service-account-id "${SAID}"
	yc resource-manager folder add-access-binding --id "${FOLDER_ID}" --role storage.admin --service-account-id "${SAID}"

delete-sa:
	yc iam service-account delete ycc-example-sa

##@ Deploy

wait-for-ycc: ## Waits for Yandex Cloud Connectors to become operational
	until [ "$(kubectl -n yandex-cloud-connectors get pod -l control-plane=connector-manager --output=json | jq '.items[0].status.phase')" = '"Running"' ]; do \
      echo "Not yet Running"; \
      sleep 1; \
    done

install: ## Installs the example setup with respect to approximate creation times
	kubectl apply -f setup/ns.yaml
	kubectl apply -f setup/sakey.yaml
	sleep 1
	kubectl apply -f setup/yos.yaml
	kubectl apply -f setup/ycr.yaml
	kubectl apply -f setup/ymq.yaml
	sleep 1
	kubectl apply -f setup/server.yaml
	kubectl apply -f setup/worker.yaml
	kubectl apply -f setup/service.yaml

uninstall:  ## Uninstalls the example setup
	kubectl delete -f setup/server.yaml
	kubectl delete -f setup/worker.yaml
	kubectl delete -f setup/service.yaml
	kubectl delete -f setup/yos.yaml
	kubectl delete -f setup/ycr.yaml
	kubectl delete -f setup/ymq.yaml
	kubectl delete -f setup/sakey.yaml
	kubectl delete -f setup/ns.yaml

##@ Running

make-sample-request: ## Makes sample request to the example service
	CLUSTER_ENDPOINT=$(kubectl -n yandex-cloud-connectors-example get service/image-reporter --output=json | jq '.status.loadBalancer.ingress[0].ip' | tr -d '"')
	curl -X POST -d "Hello Yandex Cloud Connectors!" "${CLUSTER_ENDPOINT}/report?filename=greetings.txt"