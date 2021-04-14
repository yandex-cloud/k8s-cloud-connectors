#!/bin/bash

REGISTRY="$2"

kubectl run "$1"-controller --restart=Never --rm -it --serviceaccount default --image "$REGISTRY"/"$1"-controller:latest
