// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/awskey/api/v1"
	awskeyconfig "k8s-connectors/connectors/awskey/pkg/config"
	"k8s-connectors/pkg/secrets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awskeyutils "k8s-connectors/connectors/awskey/pkg/utils"
)

type Allocator struct {
	Sdk    *ycsdk.SDK
	Client *client.Client
}

func (r *Allocator) IsUpdated(ctx context.Context, object *connectorsv1.AWSAccessKey) (bool, error) {
	res, err := awskeyutils.GetAWSAccessKey(ctx, object, r.Sdk)
	if err != nil {
		return false, err
	}
	// We do not need to check secret for existence
	// because there's no way to recreate it
	return res != nil, nil
}

func (r *Allocator) Update(ctx context.Context, log logr.Logger, object *connectorsv1.AWSAccessKey) error {
	res, err := r.Sdk.IAM().AWSCompatibility().AccessKey().Create(ctx, &awscompatibility.CreateAccessKeyRequest{
		ServiceAccountId: object.Spec.ServiceAccountID,
		Description:      awskeyutils.GetAWSAccessKeyDescription(object),
	})
	if err != nil {
		return fmt.Errorf("error while creating resource: %v", err)
	}

	// Now we need to create a secret with the key
	if err := secrets.Put(ctx, r.Client, object.ObjectMeta, awskeyconfig.ShortName, map[string]string{
		"key":    res.AccessKey.KeyId,
		"secret": res.Secret,
	}); err != nil {
		// This exact error is a disaster - we have created key, but
		// have not provided secret with secret key and therefore
		// we will inevitably lose it.
		// TODO (covariance) maybe put log.Fatal here?
		return err
	}

	// And we need to update status
	object.Status.SecretName = secrets.SecretName(object.ObjectMeta, awskeyconfig.ShortName)
	if err := (*r.Client).Update(ctx, object); err != nil {
		return fmt.Errorf("error while creating resource: %v", err)
	}

	log.Info("resource allocated successfully")
	return nil
}

func (r *Allocator) Cleanup(ctx context.Context, log logr.Logger, object *connectorsv1.AWSAccessKey) error {

	if err := awskeyutils.DeleteAWSAccessKeyAndSecret(ctx, r.Client, r.Sdk, object); err != nil {
		return err
	}

	log.Info("resource deleted successfully")
	return nil
}
