// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
)

type YandexObjectStorageAdapterSDK struct{}

func NewYandexObjectStorageAdapterSDK() (YandexObjectStorageAdapter, error) {
	return &YandexObjectStorageAdapterSDK{}, nil
}

func (r *YandexObjectStorageAdapterSDK) Create(_ context.Context, sdk *s3.S3, name string) error {
	_, err := sdk.CreateBucket(
		&s3.CreateBucketInput{
			Bucket: &name,
		},
	)
	return err
}

func (r *YandexObjectStorageAdapterSDK) List(_ context.Context, sdk *s3.S3) ([]*s3.Bucket, error) {
	res, err := sdk.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	return res.Buckets, nil
}

func (r *YandexObjectStorageAdapterSDK) Delete(_ context.Context, sdk *s3.S3, name string) error {
	_, err := sdk.DeleteBucket(
		&s3.DeleteBucketInput{
			Bucket: &name,
		},
	)
	return err
}
