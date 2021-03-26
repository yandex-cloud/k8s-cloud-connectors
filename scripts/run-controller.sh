#!/bin/bash

kubectl run custom-controller --restart=Never --rm -it --serviceaccount default --image cr.yandex/crptp7j81e7caog8r6gq/controller:latest
