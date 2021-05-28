// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	ycrutils "k8s-connectors/connector/ycr/pkg/util"
)

func (r *yandexContainerRegistryReconciler) updateStatus(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry,
) error {
	res, err := ycrutils.GetRegistry(
		ctx, object.Status.ID, object.Spec.FolderID, object.ObjectMeta.Name, r.clusterID, r.adapter,
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

	if err := r.Client.Update(ctx, object); err != nil {
		return fmt.Errorf("unable to update object status: %v", err)
	}

	log.Info("object status updated")
	return nil
}
