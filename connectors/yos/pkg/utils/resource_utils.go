// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package utils

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/errors"
)

func checkRegistryMatchWithYcr(ycr *containerregistry.Registry, registry *connectorsv1.YandexContainerRegistry) bool {
	cluster, ok1 := ycr.Labels[config.CloudClusterLabel]
	name, ok2 := ycr.Labels[config.CloudNameLabel]
	return ok1 && ok2 && cluster == registry.ClusterName && name == registry.Name
}

// getRegistryId: tries to retrieve YC ID of registry and check whether it exists
// If registry does not exist, this method returns nil
func GetRegistry(ctx context.Context, registry *connectorsv1.YandexContainerRegistry, sdk *ycsdk.SDK) (*containerregistry.Registry, error) {
	// If id is written in the status, we need to check
	// whether it exists in the cloud
	if registry.Status.Id != "" {
		ycr, err := sdk.ContainerRegistry().Registry().Get(ctx, &containerregistry.GetRegistryRequest{
			RegistryId: registry.Status.Id,
		})
		if err != nil {
			// If registry was not found then it does not exist,
			// but this error is not fatal, just a mismatch between
			// out status and real world state.
			if errors.CheckRPCErrorNotFound(err) {
				return nil, nil
			}
			// Otherwise, it is fatal
			return nil, fmt.Errorf("cannot get registry from cloud: %v", err)
		}

		// If labels do match with our object, then we have found it
		if checkRegistryMatchWithYcr(ycr, registry) {
			return ycr, nil
		}

		// Otherwise registry is not found, but that is ok:
		// we will try to list resources and find the one we need.
	}

	// TODO (covariance) pagination
	// Otherwise, we try to match cluster name and meta name
	// with registries in the cloud
	list, err := sdk.ContainerRegistry().Registry().List(ctx, &containerregistry.ListRegistriesRequest{
		FolderId: registry.Spec.FolderId,
	})
	if err != nil {
		// This error is fatal
		return nil, fmt.Errorf("cannot list registries in folder: %v", err)
	}

	for _, ycr := range list.Registries {
		// If labels do match with our object, then we have found it
		if checkRegistryMatchWithYcr(ycr, registry) {
			return ycr, nil
		}
	}

	// Nothing found, no such registry
	return nil, nil
}
