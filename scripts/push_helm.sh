#!/bin/bash
set -euxo pipefail

TAG="$TAG"
REGISTRY="$REGISTRY"
DIR="$(dirname "$0")"

CHART_IMG="$REGISTRY/chart"

cd "$DIR/.."
# required to use registries
export HELM_EXPERIMENTAL_OCI=1
# for helm to used docker auth instead of independent one
export HELM_REGISTRY_CONFIG="$HOME/.docker/config.json"
make helm-push CHART_IMG="$CHART_IMG" TAG="$TAG"
