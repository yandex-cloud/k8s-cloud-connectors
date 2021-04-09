// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package utils

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/awskey/api/v1"
	"k8s-connectors/pkg/errors"
	"k8s-connectors/pkg/secrets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awskeyconfig "k8s-connectors/connectors/awskey/pkg/config"
)

// getRegistryId: tries to retrieve YC ID of registry and check whether it exists
// If registry does not exist, this method returns nil
func GetAWSAccessKey(ctx context.Context, object *connectorsv1.AWSAccessKey, sdk *ycsdk.SDK) (*awscompatibility.AccessKey, error) {
	// If id is written in the status, we need to check
	// whether it exists in the cloud.

	if object.Status.KeyID != "" {
		res, err := sdk.IAM().AWSCompatibility().AccessKey().Get(ctx, &awscompatibility.GetAccessKeyRequest{
			AccessKeyId: object.Status.KeyID,
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
	lst, err := sdk.IAM().AWSCompatibility().AccessKey().List(ctx, &awscompatibility.ListAccessKeysRequest{
		ServiceAccountId: object.Spec.ServiceAccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list resources in cloud: %v", err)
	}

	for _, res := range lst.AccessKeys {
		if res.Description == GetAWSAccessKeyDescription(object) {
			// By description match we deduce that its our key
			return res, nil
		}
	}

	return nil, nil
}

func GetAWSAccessKeyDescription(object *connectorsv1.AWSAccessKey) string {
	return awskeyconfig.CloudClusterLabel + ":" + object.ClusterName + "\n" + awskeyconfig.CloudClusterLabel + ":" + object.Name
}

func DeleteAWSAccessKeyAndSecret(ctx context.Context, client *client.Client, sdk *ycsdk.SDK, object *connectorsv1.AWSAccessKey) error {
	if err := secrets.Remove(ctx, client, object.ObjectMeta, awskeyconfig.ShortName); err != nil {
		return err
	}

	res, err := GetAWSAccessKey(ctx, object, sdk)
	if err != nil {
		return err
	}

	if _, err := sdk.IAM().AWSCompatibility().AccessKey().Delete(ctx, &awscompatibility.DeleteAccessKeyRequest{
		AccessKeyId: res.Id,
	}); err != nil {
		return err
	}

	return nil
}
