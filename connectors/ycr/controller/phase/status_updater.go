// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controller/adapter"
	ycrutils "k8s-connectors/connectors/ycr/pkg/util"
)

type StatusUpdater struct {
	Sdk    adapter.YandexContainerRegistryAdapter
	Client *client.Client
}

func (r *StatusUpdater) IsUpdated(_ context.Context, _ logr.Logger, _ *connectorsv1.YandexContainerRegistry) (
	bool, error,
) {
	// In every reconciliation we need to update
	// status. Therefore, this updater is never
	// marked as updated.
	return false, nil
}

func (r *StatusUpdater) Update(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry,
) error {
	res, err := ycrutils.GetRegistry(
		ctx, object.Status.ID, object.Spec.FolderID, object.ObjectMeta.Name, object.ObjectMeta.ClusterName, r.Sdk,
	)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("resource not found in cloud: %v", object)
	}

	object.Status.ID = res.Id
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
