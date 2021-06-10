// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	ycrconfig "k8s-connectors/connector/ycr/pkg/config"
	ycrutils "k8s-connectors/connector/ycr/pkg/util"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/errorhandling"
)

func (r *yandexContainerRegistryReconciler) allocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry,
) (*containerregistry.Registry, error) {
	log.V(1).Info("started")

	res, err := ycrutils.GetRegistry(
		ctx, object.Status.ID, object.Spec.FolderID, object.ObjectMeta.Name, r.clusterID, r.adapter,
	)
	if err == nil {
		return res, nil
	}
	if !errorhandling.CheckConnectorErrorCode(err, ycrconfig.ErrCodeYCRNotFound) {
		return nil, fmt.Errorf("unable to get resource: %v", err)
	}

	resp, err := r.adapter.Create(
		ctx, &containerregistry.CreateRegistryRequest{
			FolderId: object.Spec.FolderID,
			Name:     object.Spec.Name,
			Labels: map[string]string{
				config.CloudClusterLabel: r.clusterID,
				config.CloudNameLabel:    object.Name,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create resource: %v", err)
	}
	log.Info("successful")
	return resp, nil
}

func (r *yandexContainerRegistryReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry,
) error {
	log.V(1).Info("started")

	ycr, err := ycrutils.GetRegistry(
		ctx, object.Status.ID, object.Spec.FolderID, object.ObjectMeta.Name, r.clusterID, r.adapter,
	)
	if err != nil {
		if errorhandling.CheckConnectorErrorCode(err, ycrconfig.ErrCodeYCRNotFound) {
			log.Info("already deleted")
			return nil
		}
		return fmt.Errorf("unable to get resource: %v", err)
	}

	if err := r.adapter.Delete(ctx, ycr.Id); err != nil {
		return fmt.Errorf("unable to delete resource: %v", err)
	}
	log.Info("successful")
	return nil
}
