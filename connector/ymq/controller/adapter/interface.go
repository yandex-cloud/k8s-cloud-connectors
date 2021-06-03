// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
)

type YandexMessageQueueAdapter interface {
	Create(ctx context.Context, key, secret string, attributes map[string]*string, queueName string) (string, error)
	GetURL(ctx context.Context, key, secret, queueName string) (string, error)
	GetAttributes(ctx context.Context, key, secret, queueURL string) (map[string]*string, error)
	List(ctx context.Context, key, secret string) ([]*string, error)
	UpdateAttributes(ctx context.Context, key, secret string, attributes map[string]*string, queueName string) error
	Delete(ctx context.Context, key, secret, queueURL string) error
}
