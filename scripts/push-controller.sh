#!/bin/bash

REGISTRY="$2"

make docker-build CONNECTOR="$1"
docker tag controller-"$1" "$REGISTRY"/"$1"-controller:latest
docker push "$REGISTRY"/controller-"$1":latest
