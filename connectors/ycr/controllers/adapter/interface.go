// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
)

type YandexContainerRegistryAdapter interface {
	Create(ctx context.Context, request *containerregistry.CreateRegistryRequest) error
	Read(ctx context.Context, registryID string, folderID string, registryName string, clusterName string) (*containerregistry.Registry, error)
	Update(ctx context.Context, registryID string, folderID string, registryName string, clusterName string, request *containerregistry.UpdateRegistryRequest) error
	Delete(ctx context.Context, registryID string, folderID string, registryName string, clusterName string, request *containerregistry.DeleteRegistryRequest) error
}
