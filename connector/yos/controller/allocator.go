// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/yos/api/v1"
	yosutils "k8s-connectors/connector/yos/pkg/util"
)

func (r *yandexObjectStorageReconciler) allocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexObjectStorage,
) error {
	log.V(1).Info("started")

	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, object, r.Client)
	if err != nil {
		return fmt.Errorf("unable to retrieve key and secret: %v", err)
	}

	lst, err := r.adapter.List(ctx, key, secret)
	if err != nil {
		return fmt.Errorf("unable to list resources: %v", err)
	}
	for _, bucket := range lst {
		if *bucket.Name == object.Name {
			log.V(1).Info("bucket found")
			return nil
		}
	}

	err = r.adapter.Create(ctx, key, secret, object.Spec.Name)
	if err != nil {
		return fmt.Errorf("unable to create resource: %v", err)
	}
	log.Info("successful")
	return nil
}

func (r *yandexObjectStorageReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexObjectStorage,
) error {
	log.V(1).Info("started")

	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, object, r.Client)
	if err != nil {
		return fmt.Errorf("unable to retrieve key and secret: %v", err)
	}

	err = r.adapter.Delete(ctx, key, secret, object.Spec.Name)
	if err != nil {
		return fmt.Errorf("unable to delete resource: %v", err)
	}

	log.Info("successful")
	return nil
}
