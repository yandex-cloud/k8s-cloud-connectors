#!/bin/bash
set -euxo pipefail

TAG="$TAG"
REGISTRY="$REGISTRY"
DIR="$(dirname "$0")"

MANAGER_IMG="$REGISTRY/manager:$TAG"
CERTIFIER_IMG="$REGISTRY/certifier:$TAG"

cd "$DIR/.."
if [[ "${1:-}" == "push" ]]; then
  make docker-push MANAGER_IMG="$MANAGER_IMG" CERTIFIER_IMG="$CERTIFIER_IMG"
else
  make docker-build MANAGER_IMG="$MANAGER_IMG" CERTIFIER_IMG="$CERTIFIER_IMG"
fi