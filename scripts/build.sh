#!/bin/bash
set -euxo pipefail

TAG="$TAG"
REGISTRY="$REGISTRY"
DIR="$(dirname "$0")"

MANAGER_IMG="$REGISTRY/manager"
CERTIFIER_IMG="$REGISTRY/certifier"

cd "$DIR/.."
if [[ "${1:-}" == "push" ]]; then
  make docker-push MANAGER_IMG="$MANAGER_IMG" CERTIFIER_IMG="$CERTIFIER_IMG" TAG="$TAG"
else
  make docker-build MANAGER_IMG="$MANAGER_IMG" CERTIFIER_IMG="$CERTIFIER_IMG" TAG="$TAG"
fi
