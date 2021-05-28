// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/yos/api/v1"
	yosconfig "k8s-connectors/connector/yos/pkg/config"
	"k8s-connectors/pkg/util"
)

func (r *yandexObjectStorageReconciler) registerFinalizer(
	ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage,
) error {
	if util.ContainsString(resource.Finalizers, yosconfig.FinalizerName) {
		return nil
	}

	resource.Finalizers = append(resource.Finalizers, yosconfig.FinalizerName)
	if err := r.Client.Update(ctx, resource); err != nil {
		return fmt.Errorf("unable to update resource status: %v", err)
	}
	log.Info("finalizer registered successfully")
	return nil
}

func (r *yandexObjectStorageReconciler) deregisterFinalizer(
	ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage,
) error {
	resource.Finalizers = util.RemoveString(resource.Finalizers, yosconfig.FinalizerName)
	if err := r.Client.Update(ctx, resource); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}

	log.Info("finalizer removed successfully")
	return nil
}
