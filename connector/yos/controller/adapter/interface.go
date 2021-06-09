// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
)

type YandexObjectStorageAdapter interface {
	Create(ctx context.Context, sdk *s3.S3, name string) error
	List(ctx context.Context, sdk *s3.S3) ([]*s3.Bucket, error)
	Delete(ctx context.Context, sdk *s3.S3, name string) error
}
