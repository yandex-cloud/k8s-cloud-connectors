#!/bin/bash

REGISTRY=cr.yandex/crptp7j81e7caog8r6gq

make docker-build CONNECTOR="$1"
docker tag controller-"$1" "$REGISTRY"/controller-"$1":latest
docker push "$REGISTRY"/controller-"$1":latest
