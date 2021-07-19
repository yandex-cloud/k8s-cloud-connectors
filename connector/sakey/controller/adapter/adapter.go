// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type StaticAccessKeyAdapterSDK struct {
	sdk *ycsdk.SDK
}

func NewStaticAccessKeyAdapter(ctx context.Context) (StaticAccessKeyAdapter, error) {
	sdk, err := ycsdk.Build(
		ctx,
		ycsdk.Config{
			Credentials: ycsdk.InstanceServiceAccount(),
		},
	)
	if err != nil {
		return nil, err
	}
	return &StaticAccessKeyAdapterSDK{
		sdk: sdk,
	}, nil
}

func (r StaticAccessKeyAdapterSDK) Create(
	ctx context.Context, saID, description string,
) (*awscompatibility.CreateAccessKeyResponse, error) {
	return r.sdk.IAM().AWSCompatibility().AccessKey().Create(
		ctx, &awscompatibility.CreateAccessKeyRequest{
			ServiceAccountId: saID,
			Description:      description,
		},
	)
}

func (r StaticAccessKeyAdapterSDK) Read(ctx context.Context, keyID string) (*awscompatibility.AccessKey, error) {
	return r.sdk.IAM().AWSCompatibility().AccessKey().Get(
		ctx, &awscompatibility.GetAccessKeyRequest{
			AccessKeyId: keyID,
		},
	)
}

func (r StaticAccessKeyAdapterSDK) Delete(ctx context.Context, sakeyID string) error {
	if _, err := r.sdk.IAM().AWSCompatibility().AccessKey().Delete(
		ctx, &awscompatibility.DeleteAccessKeyRequest{
			AccessKeyId: sakeyID,
		},
	); err != nil {
		return err
	}

	return nil
}

func (r StaticAccessKeyAdapterSDK) List(ctx context.Context, saID string) ([]*awscompatibility.AccessKey, error) {
	list, err := r.sdk.IAM().AWSCompatibility().AccessKey().List(
		ctx, &awscompatibility.ListAccessKeysRequest{
			ServiceAccountId: saID,
		},
	)
	if err != nil {
		return nil, err
	}
	return list.AccessKeys, nil
}
