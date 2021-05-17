// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrconfig "k8s-connectors/connectors/ycr/pkg/config"
	"k8s-connectors/pkg/configmap"
)

type EndpointProvider struct {
	Client client.Client
}

func (r *EndpointProvider) IsUpdated(
	ctx context.Context, _ logr.Logger, registry *connectorsv1.YandexContainerRegistry,
) (bool, error) {
	return configmap.Exists(ctx, r.Client, registry.Name, registry.Namespace, ycrconfig.ShortName)
}

func (r *EndpointProvider) Update(
	ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry,
) error {
	if err := configmap.Put(
		ctx, r.Client, registry.Name, registry.Namespace, ycrconfig.ShortName, map[string]string{
			"ID": registry.Status.ID,
		},
	); err != nil {
		return fmt.Errorf("unable to update endpoint: %v", err)
	}
	log.Info("endpoint successfully provided")
	return nil
}

func (r *EndpointProvider) Cleanup(
	ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry,
) error {
	if err := configmap.Remove(ctx, r.Client, registry.Name, registry.Namespace, ycrconfig.ShortName); err != nil {
		return fmt.Errorf("unable to remove endpoint: %v", err)
	}
	log.Info("endpoint successfully removed")
	return nil
}
