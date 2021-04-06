// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrutils "k8s-connectors/connectors/ycr/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusUpdater struct {
	Sdk    *ycsdk.SDK
	Client *client.Client
}

func (r *StatusUpdater) IsUpdated(_ context.Context, _ *connectorsv1.YandexContainerRegistry) (bool, error) {
	// In every reconciliation we need to update
	// status. Therefore, this updater is never
	//	// marked as updated.
	return false, nil
}

func (r *StatusUpdater) Update(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	ycr, err := ycrutils.GetRegistry(ctx, registry, r.Sdk)
	if err != nil {
		return fmt.Errorf("unable to get registry: %v", err)
	}
	if ycr == nil {
		return fmt.Errorf("registry %s not found in folder %s", registry.Spec.Name, registry.Spec.FolderId)
	}

	// No type check here is needed, if we cannot find it,
	// it must be something internal, otherwise getRegistryId
	// must already return error

	registry.Status.Id = ycr.Id
	// TODO (covariance) decide what to do with registry.Status.Status
	// TODO (covariance) maybe store registry.Status.CreatedAt as a timestamp?
	registry.Status.CreatedAt = ycr.CreatedAt.String()
	registry.Status.Labels = ycr.Labels

	if err := (*r.Client).Update(ctx, registry); err != nil {
		return fmt.Errorf("unable to update registry status: %v", err)
	}

	log.Info("registry status updated")
	return nil
}
