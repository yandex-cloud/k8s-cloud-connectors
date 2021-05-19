// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
)

type StaticAccessKeyPhase interface {
	IsUpdated(context.Context, logr.Logger, *connectorsv1.StaticAccessKey) (bool, error)
	Update(context.Context, logr.Logger, *connectorsv1.StaticAccessKey) error
	Cleanup(context.Context, logr.Logger, *connectorsv1.StaticAccessKey) error
}
