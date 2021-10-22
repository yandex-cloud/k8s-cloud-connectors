// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"

	ymqutil "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
)

type YandexMessageQueueAdapterSDK struct {
}

func NewYandexMessageQueueAdapterSDK() YandexMessageQueueAdapter {
	return &YandexMessageQueueAdapterSDK{}
}

func (r *YandexMessageQueueAdapterSDK) Create(
	_ context.Context, sdk *sqs.SQS, attributes map[string]*string, name string,
) (string, error) {
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

func (r *YandexMessageQueueAdapterSDK) GetURL(_ context.Context, sdk *sqs.SQS, queueName string) (string, error) {
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
	_ context.Context,
	sdk *sqs.SQS,
	queueURL string,
) (map[string]*string, error) {
	res, err := sdk.GetQueueAttributes(
		&sqs.GetQueueAttributesInput{
			AttributeNames: []*string{
				util.StringPtr(ymqutil.FifoQueue),
				util.StringPtr(ymqutil.ContentBasedDeduplication),
				util.StringPtr(ymqutil.DelaySeconds),
				util.StringPtr(ymqutil.MaximumMessageSize),
				util.StringPtr(ymqutil.MessageRetentionPeriod),
				util.StringPtr(ymqutil.ReceiveMessageWaitTimeSeconds),
				util.StringPtr(ymqutil.VisibilityTimeout),
			},
			QueueUrl: &queueURL,
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Attributes, nil
}

func (r *YandexMessageQueueAdapterSDK) List(ctx context.Context, sdk *sqs.SQS) ([]*string, error) {
	res, err := sdk.ListQueues(&sqs.ListQueuesInput{})
	if err != nil {
		return nil, err
	}

	return res.QueueUrls, nil
}

func (r *YandexMessageQueueAdapterSDK) UpdateAttributes(
	_ context.Context, sdk *sqs.SQS, attributes map[string]*string, queueURL string,
) error {
	_, err := sdk.SetQueueAttributes(
		&sqs.SetQueueAttributesInput{
			Attributes: attributes,
			QueueUrl:   &queueURL,
		},
	)
	return err
}

func (r *YandexMessageQueueAdapterSDK) Delete(_ context.Context, sdk *sqs.SQS, queueURL string) error {
	_, err := sdk.DeleteQueue(
		&sqs.DeleteQueueInput{
			QueueUrl: &queueURL,
		},
	)
	return err
}
