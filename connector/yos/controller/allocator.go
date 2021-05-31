// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/yos/api/v1"
	yosutils "k8s-connectors/connector/yos/pkg/util"
)

func (r *yandexObjectStorageReconciler) allocateResource(
	ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage,
) error {
	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, resource, r.Client)
	if err != nil {
		return err
	}

	lst, err := r.adapter.List(ctx, key, secret)
	if err != nil {
		return err
	}
	for _, bucket := range lst {
		if *bucket.Name == resource.Name {
			return nil
		}
	}

	err = r.adapter.Create(ctx, key, secret, resource.Spec.Name)
	if err != nil {
		return err
	}
	log.Info("resource successfully allocated")
	return nil
}

func (r *yandexObjectStorageReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage,
) error {
	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, resource, r.Client)
	if err != nil {
		return err
	}

	err = r.adapter.Delete(ctx, key, secret, resource.Spec.Name)
	if err != nil {
		return err
	}

	log.Info("resource successfully deleted")
	return nil
}
