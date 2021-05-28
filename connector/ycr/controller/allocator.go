// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	ycrutils "k8s-connectors/connector/ycr/pkg/util"
	"k8s-connectors/pkg/config"
)

func (r *yandexContainerRegistryReconciler) allocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry,
) error {
	res, err := ycrutils.GetRegistry(
		ctx, object.Status.ID, object.Spec.FolderID, object.ObjectMeta.Name, r.clusterID, r.adapter,
	)
	if err != nil {
		return err
	}
	if res != nil {
		return nil
	}

	if _, err := r.adapter.Create(
		ctx, &containerregistry.CreateRegistryRequest{
			FolderId: object.Spec.FolderID,
			Name:     object.Spec.Name,
			Labels: map[string]string{
				config.CloudClusterLabel: r.clusterID,
				config.CloudNameLabel:    object.Name,
			},
		},
	); err != nil {
		return err
	}
	log.Info("resource allocated successfully")
	return nil
}

func (r *yandexContainerRegistryReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry,
) error {
	ycr, err := ycrutils.GetRegistry(
		ctx, object.Status.ID, object.Spec.FolderID, object.ObjectMeta.Name, r.clusterID, r.adapter,
	)
	if err != nil {
		return err
	}
	if ycr == nil {
		log.Info("registry deleted externally")
		return nil
	}

	if err := r.adapter.Delete(ctx, ycr.Id); err != nil {
		return err
	}
	log.Info("registry deleted successfully")
	return nil
}
