// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/errors"
)

type StaticAccessKeyAdapterSDK struct {
	sdk *ycsdk.SDK
}

func NewStaticAccessKeyAdapter() (StaticAccessKeyAdapter, error) {
	sdk, err := ycsdk.Build(context.Background(), ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return nil, err
	}
	return &StaticAccessKeyAdapterSDK{
		sdk: sdk,
	}, nil
}

func getStaticAccessKeyDescription(clusterName string, name string) string {
	return config.CloudClusterLabel + ":" + clusterName + "\n" + config.CloudNameLabel + ":" + name
}

func (r StaticAccessKeyAdapterSDK) Create(ctx context.Context, saID string, clusterName string, name string) (*awscompatibility.CreateAccessKeyResponse, error) {
	return r.sdk.IAM().AWSCompatibility().AccessKey().Create(ctx, &awscompatibility.CreateAccessKeyRequest{
		ServiceAccountId: saID,
		Description:      getStaticAccessKeyDescription(clusterName, name),
	})
}

func (r StaticAccessKeyAdapterSDK) Read(ctx context.Context, keyID string, saID string, clusterName string, name string) (*awscompatibility.AccessKey, error) {
	if keyID != "" {
		res, err := r.sdk.IAM().AWSCompatibility().AccessKey().Get(ctx, &awscompatibility.GetAccessKeyRequest{
			AccessKeyId: keyID,
		})
		if err != nil {
			// If resource was not found then it does not exist,
			// but this error is not fatal, just a mismatch between
			// out status and real world state.
			if errors.CheckRPCErrorNotFound(err) {
				return nil, nil
			}
			// Otherwise, it is fatal
			return nil, fmt.Errorf("cannot get resource from cloud: %v", err)
		}

		// Everything is fine, we have found it
		return res, nil
	}

	// We may have not yet written this key into status,
	// But we can list objects and match by description
	// TODO (covariance) pagination
	lst, err := r.sdk.IAM().AWSCompatibility().AccessKey().List(ctx, &awscompatibility.ListAccessKeysRequest{
		ServiceAccountId: saID,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list resources in cloud: %v", err)
	}

	for _, res := range lst.AccessKeys {
		if res.Description == getStaticAccessKeyDescription(clusterName, name) {
			// By description match we deduce that its our key
			return res, nil
		}
	}

	return nil, nil
}

func (r StaticAccessKeyAdapterSDK) Update() error {
	return nil
}

func (r StaticAccessKeyAdapterSDK) Delete(ctx context.Context, keyID string, saID string, clusterName string, name string) error {
	res, err := r.Read(ctx, keyID, saID, clusterName, name)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}

	if _, err := r.sdk.IAM().AWSCompatibility().AccessKey().Delete(ctx, &awscompatibility.DeleteAccessKeyRequest{
		AccessKeyId: res.Id,
	}); err != nil {
		return err
	}

	return nil
}
