// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"

	"k8s-connectors/connectors/ycr/controller/adapter"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/errorhandling"
)

func checkRegistryMatchWithYcr(ycr *containerregistry.Registry, registryName, clusterName string) bool {
	cluster, ok1 := ycr.Labels[config.CloudClusterLabel]
	name, ok2 := ycr.Labels[config.CloudNameLabel]
	return ok1 && ok2 && cluster == clusterName && name == registryName
}

func GetRegistry(
	ctx context.Context, registryID, folderID, registryName, clusterName string,
	ad adapter.YandexContainerRegistryAdapter,
) (*containerregistry.Registry, error) {
	// If id is written in the status, we need to check
	// whether it exists in the cloud
	if registryID != "" {
		ycr, err := ad.Read(ctx, registryID)
		if err != nil {
			// If registry was not found then it does not exist,
			// but this error is not fatal, just a mismatch between
			// out status and real world state.
			if !errorhandling.CheckRPCErrorNotFound(err) {
				return nil, fmt.Errorf("cannot get registry from cloud: %v", err)
			}
		} else if checkRegistryMatchWithYcr(ycr, registryName, clusterName) {
			// If labels do match with our object, then we have found it
			return ycr, nil
		}
		// Otherwise registry is not found, but that is ok:
		// we will try to list resources and find the one we need.
	}

	// TODO (covariance) pagination
	list, err := ad.List(ctx, folderID)
	if err != nil {
		// This error is fatal
		return nil, fmt.Errorf("cannot list registries in folder: %v", err)
	}

	for _, res := range list {
		// If labels do match with our object, then we have found it
		if checkRegistryMatchWithYcr(res, registryName, clusterName) {
			return res, nil
		}
	}

	// Nothing found, no such registry
	return nil, nil
}
