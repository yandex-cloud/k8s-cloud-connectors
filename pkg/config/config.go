// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package config

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

const (
	CloudClusterLabel string = "managed-kubernetes-cluster-id"
	CloudNameLabel    string = "managed-kubernetes-registry-metadata-name"


	normalTimeout  = 30 * time.Second
	erroredTimeout = 30 * time.Second
)

func GetNormalResult() (ctrl.Result, error) {
	return ctrl.Result{
		RequeueAfter: normalTimeout,
	}, nil
}

func GetErroredResult(err error) (ctrl.Result, error) {
	return ctrl.Result{
		RequeueAfter: erroredTimeout,
	}, err
}

func GetNeverResult() (ctrl.Result, error) {
	return ctrl.Result{
		Requeue: false,
	}, nil
}