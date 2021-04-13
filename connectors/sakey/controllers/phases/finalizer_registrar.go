// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
	"k8s-connectors/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FinalizerRegistrar struct {
	Client *client.Client
}

func (r *FinalizerRegistrar) IsUpdated(_ context.Context, registry *connectorsv1.StaticAccessKey) (bool, error) {
	return utils.ContainsString(registry.Finalizers, sakeyconfig.FinalizerName), nil
}

func (r *FinalizerRegistrar) Update(ctx context.Context, log logr.Logger, registry *connectorsv1.StaticAccessKey) error {
	registry.Finalizers = append(registry.Finalizers, sakeyconfig.FinalizerName)
	if err := (*r.Client).Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to update status: %v", err)
	}
	log.Info("finalizer registered successfully")
	return nil
}

func (r *FinalizerRegistrar) Cleanup(ctx context.Context, log logr.Logger, registry *connectorsv1.StaticAccessKey) error {
	registry.Finalizers = utils.RemoveString(registry.Finalizers, sakeyconfig.FinalizerName)
	if err := (*r.Client).Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to remove finalizer: %v", err)
	}

	log.Info("finalizer removed successfully")
	return nil
}
