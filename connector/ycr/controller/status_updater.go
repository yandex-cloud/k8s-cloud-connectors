// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
)

func (r *yandexContainerRegistryReconciler) updateStatus(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry, res *containerregistry.Registry,
) error {
	log.V(1).Info("started")

	if object.Status.ID == res.Id &&
		object.Status.CreatedAt == res.CreatedAt.String() &&
		util.EqualsStringString(object.Status.Labels, res.Labels) {
		return nil
	}

	object.Status.ID = res.Id
	// TODO (covariance) decide what to do with object.Status.Status
	// TODO (covariance) maybe store object.Status.CreatedAt as a timestamp?
	object.Status.CreatedAt = res.CreatedAt.String()
	object.Status.Labels = res.Labels

	if err := r.Client.Update(ctx, object); err != nil {
		return fmt.Errorf("unable to update object status: %w", err)
	}

	log.Info("successful")
	return nil
}
