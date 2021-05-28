// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	ycrutils "k8s-connectors/connector/ycr/pkg/util"
)

func (r *yandexContainerRegistryReconciler) matchSpec(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry,
) error {
	log.V(1).Info("started")

	res, err := ycrutils.GetRegistry(
		ctx, object.Status.ID, object.Spec.FolderID, object.ObjectMeta.Name, r.clusterID, r.adapter,
	)
	if err != nil {
		return fmt.Errorf("unable to get resource: %v", err)
	}
	if res.Name == object.Spec.Name {
		return nil
	}

	if err := r.adapter.Update(
		ctx, &containerregistry.UpdateRegistryRequest{
			RegistryId: res.Id,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			Name:       object.Spec.Name,
		},
	); err != nil {
		return fmt.Errorf("unable to update resource: %v", err)
	}
	log.Info("successful")
	return nil
}
