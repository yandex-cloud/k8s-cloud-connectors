// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/yos/api/v1"
	yosconfig "k8s-connectors/connector/yos/pkg/config"
	"k8s-connectors/pkg/configmap"
)

func (r *yandexObjectStorageReconciler) provideConfigmap(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexObjectStorage,
) error {
	log.V(1).Info("started")

	exists, err := configmap.Exists(ctx, r.Client, object.Name, object.Namespace, yosconfig.ShortName)
	if err != nil {
		return fmt.Errorf("unable to check configmap existence: %v", err)
	}
	if exists {
		return nil
	}

	if err := configmap.Put(ctx, r.Client, object.Name, object.Namespace, yosconfig.ShortName, map[string]string{
		"name": object.Spec.Name,
	}); err != nil {
		return err
	}

	log.Info("successful")
	return nil
}

func (r *yandexObjectStorageReconciler) removeConfigMap(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexObjectStorage,
) error {
	log.V(1).Info("started")

	if err := configmap.Remove(ctx, r.Client, object.Name, object.Namespace, yosconfig.ShortName); err != nil {
		return fmt.Errorf("unable to remove configmap: %v", err)
	}
	log.Info("successful")
	return nil
}
