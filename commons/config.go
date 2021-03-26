// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package commons

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func GetErroredResult(err error) (ctrl.Result, error) {
	return ctrl.Result{
		RequeueAfter: time.Minute,
	}, err
}

func GetNormalResult() (ctrl.Result, error) {
	return ctrl.Result{
		RequeueAfter: 30 * time.Second,
	}, nil
}
