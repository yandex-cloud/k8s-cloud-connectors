// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package sdk

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
)

type YandexContainerRegistrySDK interface {
	Create(context.Context, logr.Logger, *connectorsv1.YandexContainerRegistry) error
	Read(context.Context, logr.Logger, *connectorsv1.YandexContainerRegistry) (*containerregistry.Registry, error)
	Update(context.Context, logr.Logger, *connectorsv1.YandexContainerRegistry) error
	Delete(context.Context, logr.Logger, *connectorsv1.YandexContainerRegistry) error
}
