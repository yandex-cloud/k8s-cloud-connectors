// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrconfig "k8s-connectors/connectors/ycr/pkg/config"
	"k8s-connectors/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FinalizerRegistrar struct {
	Client *client.Client
}

func (r *FinalizerRegistrar) IsUpdated(_ context.Context, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return utils.ContainsString(registry.Finalizers, ycrconfig.FinalizerName), nil
}

func (r *FinalizerRegistrar) Update(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	registry.Finalizers = append(registry.Finalizers, ycrconfig.FinalizerName)
	if err := (*r.Client).Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to update registry status: %v", err)
	}
	log.Info("finalizer registered successfully")
	return nil
}

func (r *FinalizerRegistrar) Cleanup(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	registry.Finalizers = utils.RemoveString(registry.Finalizers, ycrconfig.FinalizerName)
	if err := (*r.Client).Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}

	log.Info("finalizer removed successfully")
	return nil
}