// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"

	"k8s-connectors/connectors/ymq/pkg/utils"
)

type YandexMessageQueueAdapterSDK struct {
	sqsProvider utils.AwsSdkProvider
}

func NewYandexMessageQueueAdapterSDK() (YandexMessageQueueAdapter, error) {
	return &YandexMessageQueueAdapterSDK{
		sqsProvider: utils.NewStaticProvider(),
	}, nil
}

func (r YandexMessageQueueAdapterSDK) Create(ctx context.Context, key, secret string, attributes map[string]*string, name string) (string, error) {
	sdk, err := r.sqsProvider(ctx, key, secret)
	if err != nil {
		return "", err
	}

	res, err := sdk.CreateQueue(&sqs.CreateQueueInput{
		Attributes: attributes,
		QueueName:  &name,
	})
	if err != nil {
		return "", err
	}

	return *res.QueueUrl, nil
}

func (r YandexMessageQueueAdapterSDK) List(ctx context.Context, key, secret string) ([]*string, error) {
	sdk, err := r.sqsProvider(ctx, key, secret)
	if err != nil {
		return nil, err
	}

	res, err := sdk.ListQueues(&sqs.ListQueuesInput{})
	if err != nil {
		return nil, err
	}

	return res.QueueUrls, nil
}

func (r YandexMessageQueueAdapterSDK) Update() error {
	return nil
}

func (r YandexMessageQueueAdapterSDK) Delete(ctx context.Context, key, secret, queueURL string) error {
	sdk, err := r.sqsProvider(ctx, key, secret)
	if err != nil {
		return err
	}

	_, err = sdk.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: &queueURL,
	})
	return err
}
