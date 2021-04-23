// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

type StaticAccessKeyAdapter interface {
	Create(ctx context.Context, saID string, clusterName string, name string) (*awscompatibility.CreateAccessKeyResponse, error)
	Read(ctx context.Context, keyID string, saID string, clusterName string, name string) (*awscompatibility.AccessKey, error)
	Update() error
	Delete(ctx context.Context, keyID string, saID string, clusterName string, name string) error
}
