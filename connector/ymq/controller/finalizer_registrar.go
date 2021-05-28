// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/ymq/api/v1"
	ymqconfig "k8s-connectors/connector/ymq/pkg/config"
	"k8s-connectors/pkg/util"
)

func (r *yandexMessageQueueReconciler) registerFinalizer(
	ctx context.Context, log logr.Logger, resource *connectorsv1.YandexMessageQueue,
) error {
	if util.ContainsString(resource.Finalizers, ymqconfig.FinalizerName) {
		return nil
	}

	resource.Finalizers = append(resource.Finalizers, ymqconfig.FinalizerName)
	if err := r.Client.Update(ctx, resource); err != nil {
		return fmt.Errorf("unable to update resource status: %v", err)
	}
	log.Info("finalizer registered successfully")
	return nil
}

func (r *yandexMessageQueueReconciler) deregisterFinalizer(
	ctx context.Context, log logr.Logger, resource *connectorsv1.YandexMessageQueue,
) error {
	resource.Finalizers = util.RemoveString(resource.Finalizers, ymqconfig.FinalizerName)
	if err := r.Client.Update(ctx, resource); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}

	log.Info("finalizer removed successfully")
	return nil
}
