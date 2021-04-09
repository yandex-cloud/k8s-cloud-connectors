// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package utils

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/errors"
	"k8s-connectors/pkg/secrets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
)

// GetStaticAccessKey: tries to retrieve YC resource and check whether it exists.
// If resource does not exist or cannot be found, this method returns nil.
func GetStaticAccessKey(ctx context.Context, object *connectorsv1.StaticAccessKey, sdk *ycsdk.SDK) (*awscompatibility.AccessKey, error) {
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
		if res.Description == GetStaticAccessKeyDescription(object) {
			// By description match we deduce that its our key
			return res, nil
		}
	}

	return nil, nil
}

func GetStaticAccessKeyDescription(object *connectorsv1.StaticAccessKey) string {
	return config.CloudClusterLabel + ":" + object.ClusterName + "\n" + config.CloudClusterLabel + ":" + object.Name
}

func DeleteStaticAccessKeyAndSecret(ctx context.Context, client *client.Client, sdk *ycsdk.SDK, object *connectorsv1.StaticAccessKey) error {
	if err := secrets.Remove(ctx, client, object.ObjectMeta, sakeyconfig.ShortName); err != nil {
		return err
	}

	res, err := GetStaticAccessKey(ctx, object, sdk)
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
