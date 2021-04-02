// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package config

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

const (
	erroredTimeout = 30 * time.Second
	normalTimeout = 30 * time.Second
)

func GetErroredResult(err error) (ctrl.Result, error) {
	return ctrl.Result{
		RequeueAfter: erroredTimeout,
	}, err
}

func GetNormalResult() (ctrl.Result, error) {
	return ctrl.Result{
		RequeueAfter: 30 * normalTimeout,
	}, nil
}
