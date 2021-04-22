// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/yos/api/v1"
	yosutils "k8s-connectors/connectors/yos/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceAllocator struct {
	Client     *client.Client
	S3provider yosutils.AwsSdkProvider
}

func (r *ResourceAllocator) IsUpdated(ctx context.Context, resource *connectorsv1.YandexObjectStorage) (bool, error) {
	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return false, err
	}
	sdk, err := r.S3provider(ctx, key, secret)
	if err != nil {
		return false, err
	}

	res, err := sdk.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return false, fmt.Errorf("cannot list resources in cloud: %v", err)
	}

	for _, bucket := range res.Buckets {
		if *bucket.Name == resource.Spec.Name {
			return true, nil
		}
	}

	return false, nil
}

func (r *ResourceAllocator) Update(ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage) error {
	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return err
	}
	sdk, err := r.S3provider(ctx, key, secret)
	if err != nil {
		return err
	}

	if _, err = sdk.CreateBucket(&s3.CreateBucketInput{
		Bucket: &resource.Spec.Name,
	}); err != nil {
		return fmt.Errorf("error while creating resource: %v", err)
	}

	log.Info("resource successfully allocated")
	return nil
}

func (r *ResourceAllocator) Cleanup(ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage) error {
	key, secret, err := yosutils.KeyAndSecretFromStaticAccessKey(ctx, resource, *r.Client)
	if err != nil {
		return err
	}
	sdk, err := r.S3provider(ctx, key, secret)
	if err != nil {
		return err
	}

	if _, err = sdk.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &resource.Spec.Name,
	}); err != nil {
		return fmt.Errorf("error while deleting resource: %v", err)
	}

	log.Info("resource successfully deleted")
	return nil
}
