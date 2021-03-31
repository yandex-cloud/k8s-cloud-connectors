#!/bin/bash

REGISTRY="$2"

make docker-build CONNECTOR="$1"
docker tag controller-"$1" "$REGISTRY"/controller-"$1":latest
docker push "$REGISTRY"/controller-"$1":latest
