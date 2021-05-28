// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	ycrconfig "k8s-connectors/connector/ycr/pkg/config"
	"k8s-connectors/pkg/util"
)

func (r *yandexContainerRegistryReconciler) registerFinalizer(
	ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry,
) error {
	if util.ContainsString(registry.Finalizers, ycrconfig.FinalizerName) {
		return nil
	}
	registry.Finalizers = append(registry.Finalizers, ycrconfig.FinalizerName)
	if err := r.Client.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to update registry status: %v", err)
	}
	log.Info("finalizer registered successfully")
	return nil
}

func (r *yandexContainerRegistryReconciler) deregisterFinalizer(
	ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry,
) error {
	registry.Finalizers = util.RemoveString(registry.Finalizers, ycrconfig.FinalizerName)
	if err := r.Client.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}

	log.Info("finalizer removed successfully")
	return nil
}
