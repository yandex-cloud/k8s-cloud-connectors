// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrconfig "k8s-connectors/connectors/ycr/pkg/config"
	"k8s-connectors/pkg/configmaps"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EndpointProvider struct {
	Client *client.Client
}

func (r *EndpointProvider) IsUpdated(ctx context.Context, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	return configmaps.Exists(ctx, r.Client, registry.Name, registry.Namespace, "ycr")
}

func (r *EndpointProvider) Update(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	if err := configmaps.Put(ctx, r.Client, registry.Name, registry.Namespace, "ycr", map[string]string{
		"ID": registry.Status.Id,
	}); err != nil {
		return fmt.Errorf("unable to update endpoint: %v", err)
	}
	log.Info("endpoint successfully provided")
	return nil
}

func (r *EndpointProvider) Cleanup(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	if err := configmaps.Remove(ctx, *r.Client, registry.Name, registry.Namespace, ycrconfig.ShortName); err != nil {
		return fmt.Errorf("unable to remove endpoint: %v", err)
	}
	log.Info("endpoint successfully removed")
	return nil
}
