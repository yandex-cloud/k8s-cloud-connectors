// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connectors/yos/api/v1"
)

type YandexObjectStoragePhase interface {
	IsUpdated(context.Context, *connectorsv1.YandexObjectStorage) (bool, error)
	Update(context.Context, logr.Logger, *connectorsv1.YandexObjectStorage) error
	Cleanup(context.Context, logr.Logger, *connectorsv1.YandexObjectStorage) error
}
