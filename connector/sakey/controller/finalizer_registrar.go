// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/sakey/api/v1"
	sakeyconfig "k8s-connectors/connector/sakey/pkg/config"
	"k8s-connectors/pkg/util"
)

func (r *staticAccessKeyReconciler) registerFinalizer(
	ctx context.Context, log logr.Logger, registry *connectorsv1.StaticAccessKey,
) error {
	if util.ContainsString(registry.Finalizers, sakeyconfig.FinalizerName) {
		return nil
	}
	registry.Finalizers = append(registry.Finalizers, sakeyconfig.FinalizerName)
	if err := r.Client.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to update status: %v", err)
	}
	log.Info("finalizer registered successfully")
	return nil
}

func (r *staticAccessKeyReconciler) deregisterFinalizer(
	ctx context.Context, log logr.Logger, registry *connectorsv1.StaticAccessKey,
) error {
	registry.Finalizers = util.RemoveString(registry.Finalizers, sakeyconfig.FinalizerName)
	if err := r.Client.Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}

	log.Info("finalizer removed successfully")
	return nil
}
