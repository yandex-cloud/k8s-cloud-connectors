// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"

	"k8s-connectors/connector/ymq/pkg/util"
)

type YandexMessageQueueAdapterSDK struct {
	sqsProvider util.SQSProvider
}

func NewYandexMessageQueueAdapterSDK() (YandexMessageQueueAdapter, error) {
	return &YandexMessageQueueAdapterSDK{
		sqsProvider: util.NewStaticProvider(),
	}, nil
}

func (r *YandexMessageQueueAdapterSDK) Create(
	ctx context.Context, key, secret string, attributes map[string]*string, name string,
) (string, error) {
	sdk, err := r.sqsProvider(ctx, key, secret)
	if err != nil {
		return "", err
	}

	res, err := sdk.CreateQueue(
		&sqs.CreateQueueInput{
			Attributes: attributes,
			QueueName:  &name,
		},
	)
	if err != nil {
		return "", err
	}

	return *res.QueueUrl, nil
}

func (r *YandexMessageQueueAdapterSDK) GetURL(ctx context.Context, key, secret, queueName string) (string, error) {
	sdk, err := r.sqsProvider(ctx, key, secret)
	if err != nil {
		return "", err
	}

	res, err := sdk.GetQueueUrl(
		&sqs.GetQueueUrlInput{
			QueueName:              &queueName,
			QueueOwnerAWSAccountId: nil,
		},
	)
	if err != nil {
		return "", err
	}

	return *res.QueueUrl, nil
}

func (r *YandexMessageQueueAdapterSDK) GetAttributes(
	ctx context.Context, key, secret, queueURL string,
) (map[string]*string, error) {
	sdk, err := r.sqsProvider(ctx, key, secret)
	if err != nil {
		return nil, err
	}

	fifoQueue := "FifoQueue"
	contentBasedDeduplication := "ContentBasedDeduplication"
	delaySeconds := "DelaySeconds"
	maximumMessageSize := "MaximumMessageSize"
	messageRetentionPeriod := "MessageRetentionPeriod"
	receiveMessageWaitTimeSeconds := "ReceiveMessageWaitTimeSeconds"
	visibilityTimeout := "VisibilityTimeout"

	attributes := []*string{
		&fifoQueue,
		&contentBasedDeduplication,
		&delaySeconds,
		&maximumMessageSize,
		&messageRetentionPeriod,
		&receiveMessageWaitTimeSeconds,
		&visibilityTimeout,
	}

	res, err := sdk.GetQueueAttributes(
		&sqs.GetQueueAttributesInput{
			AttributeNames: attributes,
			QueueUrl:       &queueURL,
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Attributes, nil
}

func (r *YandexMessageQueueAdapterSDK) List(ctx context.Context, key, secret string) ([]*string, error) {
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

func (r *YandexMessageQueueAdapterSDK) UpdateAttributes(
	ctx context.Context, key, secret string, attributes map[string]*string, queueURL string,
) error {
	sdk, err := r.sqsProvider(ctx, key, secret)
	if err != nil {
		return err
	}

	_, err = sdk.SetQueueAttributes(
		&sqs.SetQueueAttributesInput{
			Attributes: attributes,
			QueueUrl:   &queueURL,
		},
	)
	return err
}

func (r *YandexMessageQueueAdapterSDK) Delete(ctx context.Context, key, secret, queueURL string) error {
	sdk, err := r.sqsProvider(ctx, key, secret)
	if err != nil {
		return err
	}

	_, err = sdk.DeleteQueue(
		&sqs.DeleteQueueInput{
			QueueUrl: &queueURL,
		},
	)
	return err
}
