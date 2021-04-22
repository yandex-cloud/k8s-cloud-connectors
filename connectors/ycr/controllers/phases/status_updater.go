// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/sdk"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusUpdater struct {
	Sdk    sdk.YandexContainerRegistrySDK
	Client *client.Client
}

func (r *StatusUpdater) IsUpdated(_ context.Context, _ logr.Logger, _ *connectorsv1.YandexContainerRegistry) (bool, error) {
	// In every reconciliation we need to update
	// status. Therefore, this updater is never
	// marked as updated.
	return false, nil
}

func (r *StatusUpdater) Update(ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry) error {
	res, err := r.Sdk.Read(ctx, log, object)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("resource not found in cloud: %v", object)
	}

	object.Status.Id = res.Id
	// TODO (covariance) decide what to do with object.Status.Status
	// TODO (covariance) maybe store object.Status.CreatedAt as a timestamp?
	object.Status.CreatedAt = res.CreatedAt.String()
	object.Status.Labels = res.Labels

	if err := (*r.Client).Update(ctx, object); err != nil {
		return fmt.Errorf("unable to update object status: %v", err)
	}

	log.Info("object status updated")
	return nil
}

func (r *StatusUpdater) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.YandexContainerRegistry) error {
	return nil
}
