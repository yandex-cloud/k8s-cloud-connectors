// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	ycrconfig "k8s-connectors/connector/ycr/pkg/config"
	"k8s-connectors/pkg/configmap"
)

func (r *yandexContainerRegistryReconciler) provideConfigMap(
	ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry,
) error {
	log.V(1).Info("started")

	exists, err := configmap.Exists(ctx, r.Client, registry.Name, registry.Namespace, ycrconfig.ShortName)
	if err != nil {
		return fmt.Errorf("unable to check configmap existance: %v", err)
	}
	if exists {
		return nil
	}

	if err := configmap.Put(
		ctx, r.Client, registry.Name, registry.Namespace, ycrconfig.ShortName, map[string]string{
			"ID": registry.Status.ID,
		},
	); err != nil {
		return fmt.Errorf("unable to update configmap: %v", err)
	}
	log.Info("successful")
	return nil
}

func (r *yandexContainerRegistryReconciler) removeConfigMap(
	ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry,
) error {
	log.V(1).Info("started")

	if err := configmap.Remove(ctx, r.Client, registry.Name, registry.Namespace, ycrconfig.ShortName); err != nil {
		return fmt.Errorf("unable to remove configmap: %v", err)
	}
	log.Info("successful")
	return nil
}
