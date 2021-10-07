#!/bin/bash
set -euxo pipefail

TAG="$TAG"
REGISTRY="$REGISTRY"
DIR="$(dirname "$0")"

CHART_IMG="$REGISTRY/chart"

cd "$DIR/.."
sed -i "s/0.0.0/$TAG/g" helm/yandex-cloud-connectors/Chart.yaml
sed -i "s/latest/$TAG/g" helm/yandex-cloud-connectors/Chart.yaml
# required to use registries
export HELM_EXPERIMENTAL_OCI=1
# for helm to used docker auth instead of independent one
export HELM_REGISTRY_CONFIG="$HOME/.docker/config.json"
make helm-push CHART_IMG="$CHART_IMG" TAG="$TAG"
