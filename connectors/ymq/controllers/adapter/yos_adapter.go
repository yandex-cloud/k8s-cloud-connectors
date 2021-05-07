// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
	"k8s-connectors/connectors/yos/pkg/utils"
)

type YandexMessageQueueAdapterSDK struct {
	s3provider utils.AwsSdkProvider
}

func NewYandexMessageQueueAdapterSDK() (YandexMessageQueueAdapter, error) {
	return &YandexMessageQueueAdapterSDK{
		s3provider: utils.NewStaticProvider(),
	}, nil
}

func (r YandexMessageQueueAdapterSDK) Create(ctx context.Context, key string, secret string, name string) error {
	sdk, err := r.s3provider(ctx, key, secret)
	if err != nil {
		return err
	}

	_, err = sdk.CreateBucket(&s3.CreateBucketInput{
		Bucket: &name,
	})
	return err
}

func (r YandexMessageQueueAdapterSDK) List(ctx context.Context, key string, secret string) ([]*s3.Bucket, error) {
	sdk, err := r.s3provider(ctx, key, secret)
	if err != nil {
		return nil, err
	}

	res, err := sdk.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	return res.Buckets, nil
}

func (r YandexMessageQueueAdapterSDK) Update() error {
	return nil
}
func (r YandexMessageQueueAdapterSDK) Delete(ctx context.Context, key string, secret string, name string) error {
	sdk, err := r.s3provider(ctx, key, secret)
	if err != nil {
		return err
	}

	_, err = sdk.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &name,
	})
	return err
}
