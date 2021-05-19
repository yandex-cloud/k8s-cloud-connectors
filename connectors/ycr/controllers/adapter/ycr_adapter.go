// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type YandexContainerRegistryAdapterSDK struct {
	sdk *ycsdk.SDK
}

func NewYandexContainerRegistryAdapterSDK() (YandexContainerRegistryAdapter, error) {
	sdk, err := ycsdk.Build(
		context.Background(), ycsdk.Config{
			Credentials: ycsdk.InstanceServiceAccount(),
		},
	)

	if err != nil {
		return nil, err
	}
	return YandexContainerRegistryAdapterSDK{
		sdk: sdk,
	}, nil
}

func (r YandexContainerRegistryAdapterSDK) Create(
	ctx context.Context, request *containerregistry.CreateRegistryRequest,
) (*containerregistry.Registry, error) {
	op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Create(ctx, request))

	if err != nil {
		return nil, err
	}

	if err = op.Wait(ctx); err != nil {
		return nil, err
	}

	res, err := op.Response()
	if err != nil {
		return nil, err
	}

	return res.(*containerregistry.Registry), nil
}

func (r YandexContainerRegistryAdapterSDK) Read(ctx context.Context, registryID string) (
	*containerregistry.Registry, error,
) {
	return r.sdk.ContainerRegistry().Registry().Get(
		ctx, &containerregistry.GetRegistryRequest{
			RegistryId: registryID,
		},
	)
}

func (r YandexContainerRegistryAdapterSDK) List(ctx context.Context, folderID string) (
	[]*containerregistry.Registry, error,
) {
	list, err := r.sdk.ContainerRegistry().Registry().List(
		ctx, &containerregistry.ListRegistriesRequest{
			FolderId: folderID,
		},
	)
	if err != nil {
		return nil, err
	}
	return list.Registries, nil
}

func (r YandexContainerRegistryAdapterSDK) Update(
	ctx context.Context, request *containerregistry.UpdateRegistryRequest,
) error {
	op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Update(ctx, request))
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

func (r YandexContainerRegistryAdapterSDK) Delete(ctx context.Context, registryID string) error {
	op, err := r.sdk.WrapOperation(
		r.sdk.ContainerRegistry().Registry().Delete(
			ctx, &containerregistry.DeleteRegistryRequest{
				RegistryId: registryID,
			},
		),
	)
	if err != nil {
		return err
	}
	if err := op.Wait(ctx); err != nil {
		return err
	}
	return nil
}
