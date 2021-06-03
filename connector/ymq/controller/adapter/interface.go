// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
)

type YandexMessageQueueAdapter interface {
	Create(ctx context.Context, key, secret string, attributes map[string]*string, name string) (string, error)
	List(ctx context.Context, key, secret string) ([]*string, error)
	Delete(ctx context.Context, key, secret, queueURL string) error
}
