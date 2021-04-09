// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
)

type YandexContainerRegistryPhase interface {
	IsUpdated(context.Context, *connectorsv1.YandexContainerRegistry) (bool, error)
	Update(context.Context, logr.Logger, *connectorsv1.YandexContainerRegistry) error
	Cleanup(context.Context, logr.Logger, *connectorsv1.YandexContainerRegistry) error
}
