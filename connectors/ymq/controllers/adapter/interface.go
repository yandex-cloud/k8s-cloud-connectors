// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
)

type YandexMessageQueueAdapter interface {
	Create(ctx context.Context, key string, secret string, name string) error
	List(ctx context.Context, key string, secret string) ([]*s3.Bucket, error)
	Delete(ctx context.Context, key string, secret string, name string) error
}
