// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/api/v1"
	ycrconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/pkg/config"
	ycrutils "github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/config"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/errorhandling"
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
		return nil, fmt.Errorf("unable to get resource: %w", err)
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
		return nil, fmt.Errorf("unable to create resource: %w", err)
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
		return fmt.Errorf("unable to get resource: %w", err)
	}

	if err := r.adapter.Delete(ctx, ycr.Id); err != nil {
		return fmt.Errorf("unable to delete resource: %w", err)
	}
	log.Info("successful")
	return nil
}
