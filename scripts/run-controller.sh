#!/bin/bash

REGISTRY="$2"

kubectl run controller-"$1" --restart=Never --rm -it --serviceaccount default --image "$REGISTRY"/controller-"$1":latest
