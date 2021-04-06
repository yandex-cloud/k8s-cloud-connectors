// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/pkg/configmaps"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EntrypointProvider struct {
	Client *client.Client
}

func (r *EntrypointProvider) IsUpdated(_ context.Context, _ *connectorsv1.YandexContainerRegistry) (bool, error) {
	// In every reconciliation we need to update
	// endpoint. Therefore, this updater is never
	// marked as updated.
	return false, nil
}

func (r *EntrypointProvider) Update(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	if err := configmaps.Put(ctx, r.Client, registry, map[string]string{
		"ID": registry.Status.Id,
	}); err != nil {
		return fmt.Errorf("unable to update entrypoint: %v", err)
	}
	log.Info("entrypoint updated")
	return nil
}
