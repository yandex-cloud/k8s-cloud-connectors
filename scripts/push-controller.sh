#!/bin/bash

REGISTRY="$2"

make docker-build CONNECTOR="$1"
docker tag "$1"-controller "$REGISTRY"/"$1"-controller:latest
docker push "$REGISTRY"/"$1"-controller:latest
