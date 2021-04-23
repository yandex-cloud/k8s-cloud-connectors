// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/yos/api/v1"
	"k8s-connectors/connectors/yos/controllers/adapter"
	yosutils "k8s-connectors/connectors/yos/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceAllocator struct {
	Client *client.Client
	Sdk    adapter.YandexObjectStorageAdapter
}

func (r *ResourceAllocator) IsUpdated(ctx context.Context, resource *connectorsv1.YandexObjectStorage) (bool, error) {
	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return false, err
	}
	res, err := r.Sdk.Read(ctx, key, secret, resource.Name)
	return res != nil, err
}

func (r *ResourceAllocator) Update(ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage) error {
	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return err
	}
	err = r.Sdk.Create(ctx, key, secret, resource.Spec.Name)
	if err != nil {
		return err
	}
	log.Info("resource successfully allocated")
	return nil
}

func (r *ResourceAllocator) Cleanup(ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage) error {
	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return err
	}

	err = r.Sdk.Delete(ctx, key, secret, resource.Spec.Name)
	if err != nil {
		return err
	}

	log.Info("resource successfully deleted")
	return nil
}
