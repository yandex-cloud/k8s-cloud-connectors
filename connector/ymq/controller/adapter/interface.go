// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type YandexMessageQueueAdapter interface {
	Create(ctx context.Context, sdk *sqs.SQS, attributes map[string]*string, queueName string) (string, error)
	GetURL(ctx context.Context, sdk *sqs.SQS, queueName string) (string, error)
	GetAttributes(ctx context.Context, sdk *sqs.SQS, queueURL string) (map[string]*string, error)
	List(ctx context.Context, sdk *sqs.SQS) ([]*string, error)
	UpdateAttributes(ctx context.Context, sdk *sqs.SQS, attributes map[string]*string, queueName string) error
	Delete(ctx context.Context, sdk *sqs.SQS, queueURL string) error
}
