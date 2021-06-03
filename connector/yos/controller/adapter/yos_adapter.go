// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"

	"k8s-connectors/connector/yos/pkg/util"
)

type YandexObjectStorageAdapterSDK struct {
	s3provider util.S3Provider
}

func NewYandexObjectStorageAdapterSDK() (YandexObjectStorageAdapter, error) {
	return &YandexObjectStorageAdapterSDK{
		s3provider: util.NewStaticProvider(),
	}, nil
}

func (r YandexObjectStorageAdapterSDK) Create(ctx context.Context, key, secret, name string) error {
	sdk, err := r.s3provider(ctx, key, secret)
	if err != nil {
		return err
	}

	_, err = sdk.CreateBucket(
		&s3.CreateBucketInput{
			Bucket: &name,
		},
	)
	return err
}

func (r YandexObjectStorageAdapterSDK) List(ctx context.Context, key, secret string) ([]*s3.Bucket, error) {
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

func (r YandexObjectStorageAdapterSDK) Update() error {
	return nil
}

func (r YandexObjectStorageAdapterSDK) Delete(ctx context.Context, key, secret, name string) error {
	sdk, err := r.s3provider(ctx, key, secret)
	if err != nil {
		return err
	}

	_, err = sdk.DeleteBucket(
		&s3.DeleteBucketInput{
			Bucket: &name,
		},
	)
	return err
}
