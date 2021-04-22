// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/sdk"
)

type Allocator struct {
	Sdk sdk.YandexContainerRegistrySDK
}

func (r *Allocator) IsUpdated(ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry) (bool, error) {
	res, err := r.Sdk.Read(ctx, log, object)
	return res != nil, err
}

func (r *Allocator) Update(ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry) error {
	if err := r.Sdk.Create(ctx, log, object); err != nil {
		return err
	}
	log.Info("resource allocated successfully")
	return nil
}

func (r *Allocator) Cleanup(ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry) error {
	if err := r.Sdk.Delete(ctx, log, object); err != nil {
		return err
	}
	log.Info("registry deleted successfully")
	return nil
}
