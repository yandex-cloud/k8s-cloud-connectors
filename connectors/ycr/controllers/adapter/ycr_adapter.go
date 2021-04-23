// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/errors"
)

type YandexContainerRegistryAdapterSDK struct {
	sdk *ycsdk.SDK
}

func NewYandexContainerRegistryAdapterSDK() (YandexContainerRegistryAdapter, error) {
	sdk, err := ycsdk.Build(context.Background(), ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})

	if err != nil {
		return nil, err
	}
	return YandexContainerRegistryAdapterSDK{
		sdk: sdk,
	}, nil
}

func (r YandexContainerRegistryAdapterSDK) Create(ctx context.Context, request *containerregistry.CreateRegistryRequest) error {
	op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Create(ctx, request))

	if err != nil {
		return err
	}

	if err := op.Wait(ctx); err != nil {
		return err
	}

	if _, err := op.Response(); err != nil {
		return err
	}

	return nil
}

func checkRegistryMatchWithYcr(ycr *containerregistry.Registry, registryName string, clusterName string) bool {
	cluster, ok1 := ycr.Labels[config.CloudClusterLabel]
	name, ok2 := ycr.Labels[config.CloudNameLabel]
	return ok1 && ok2 && cluster == clusterName && name == registryName
}

func (r YandexContainerRegistryAdapterSDK) Read(ctx context.Context, registryID string, folderID string, registryName string, clusterName string) (*containerregistry.Registry, error) {
	// If id is written in the status, we need to check
	// whether it exists in the cloud
	if registryID != "" {
		ycr, err := r.sdk.ContainerRegistry().Registry().Get(ctx, &containerregistry.GetRegistryRequest{
			RegistryId: registryID,
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
		if checkRegistryMatchWithYcr(ycr, registryName, clusterName) {
			return ycr, nil
		}

		// Otherwise registry is not found, but that is ok:
		// we will try to list resources and find the one we need.
	}

	// TODO (covariance) pagination
	list, err := r.sdk.ContainerRegistry().Registry().List(ctx, &containerregistry.ListRegistriesRequest{
		FolderId: folderID,
	})
	if err != nil {
		// This error is fatal
		return nil, fmt.Errorf("cannot list registries in folder: %v", err)
	}

	for _, ycr := range list.Registries {
		// If labels do match with our object, then we have found it
		if checkRegistryMatchWithYcr(ycr, registryName, clusterName) {
			return ycr, nil
		}
	}

	// Nothing found, no such registry
	return nil, nil
}

func (r YandexContainerRegistryAdapterSDK) Update(ctx context.Context,  registryID string, folderID string, registryName string, clusterName string, request *containerregistry.UpdateRegistryRequest) error {
	res, err := r.Read(ctx, registryID, folderID, registryName, clusterName)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("cannot update registry")
	}

	op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Update(ctx, request))

	if err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}

	return nil
}

func (r YandexContainerRegistryAdapterSDK) Delete(ctx context.Context, registryID string, folderID string, registryName string, clusterName string, request *containerregistry.DeleteRegistryRequest) error {
	ycr, err := r.Read(ctx, registryID, folderID, registryName, clusterName)
	if err != nil {
		return err
	}

	if ycr != nil {
		op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Delete(ctx, request))
		if err != nil {
			return err
		}
		if err := op.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}
