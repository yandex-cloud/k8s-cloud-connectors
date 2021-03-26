#!/bin/bash
# это коммент
make docker-build
docker tag controller cr.yandex/crptp7j81e7caog8r6gq/controller:latest
docker push cr.yandex/crptp7j81e7caog8r6gq/controller:latest
