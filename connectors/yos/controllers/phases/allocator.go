// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-logr/logr"
	sakey "k8s-connectors/connectors/sakey/api/v1"
	connectorsv1 "k8s-connectors/connectors/yos/api/v1"
	yosutils "k8s-connectors/connectors/yos/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceAllocator struct {
	Client     *client.Client
	S3provider yosutils.AwsSdkProvider
}

func (r *ResourceAllocator) IsUpdated(_ context.Context, resource *connectorsv1.YandexObjectStorage) (bool, error) {
	return resource.Status.Created, nil
}

func (r *ResourceAllocator) Update(ctx context.Context, log logr.Logger, resource *connectorsv1.YandexObjectStorage) error {
	var key sakey.StaticAccessKey
	if err := (*r.Client).Get(ctx, types.NamespacedName{
		Namespace: "default",
		Name: resource.Spec.SAKeyRef,
	}, &key); err != nil {
		return fmt.Errorf("unable to retrieve corresponding SAKey: %v", err)
	}

	var secret v1.Secret
	if err := (*r.Client).Get(ctx, types.NamespacedName{
		Namespace: "default",
		Name:      key.Status.SecretName,
	}, &secret); err != nil {
		return fmt.Errorf("unable to retrieve corresponding secret: %v", err)
	}

	log.Info(secret.StringData["key"], secret.StringData["secret"])

	sdk, err := r.S3provider(ctx, secret.StringData["key"], secret.StringData["secret"])
	if err != nil {
		return err
	}

	if _, err = sdk.CreateBucket(&s3.CreateBucketInput{
		Bucket:                     &resource.Spec.Name,
		GrantRead:                  &resource.Spec.ReadAccess,
		GrantReadACP:               &resource.Spec.ReadAccess,
	}); err != nil {
		return fmt.Errorf("error while creating bucket: %v", err)
	}

	return nil
}

func (r *ResourceAllocator) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.YandexObjectStorage) error {
	return nil
}
