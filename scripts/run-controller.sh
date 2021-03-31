#!/bin/bash

REGISTRY=cr.yandex/crptp7j81e7caog8r6gq

kubectl run controller-"$1" --restart=Never --rm -it --serviceaccount default --image "$REGISTRY"/controller-"$1":latest
