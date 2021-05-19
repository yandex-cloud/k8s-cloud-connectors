// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

type StaticAccessKeyAdapter interface {
	Create(ctx context.Context, saID string, description string) (*awscompatibility.CreateAccessKeyResponse, error)
	Read(ctx context.Context, keyID string) (*awscompatibility.AccessKey, error)
	Delete(ctx context.Context, sakeyID string) error
	List(ctx context.Context, saID string) ([]*awscompatibility.AccessKey, error)
}
