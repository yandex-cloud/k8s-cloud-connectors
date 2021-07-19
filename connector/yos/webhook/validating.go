// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"

	sakey "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/api/v1"
	yosutils "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/awsutils"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/yos/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexobjectstorage,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexobjectstorages,verbs=create;update;delete,versions=v1,name=vyandexobjectstorage.yandex.com,admissionReviewVersions=v1

type YOSValidator struct {
	cl client.Client
}

func NewYOSValidator(cl client.Client) (webhook.Validator, error) {
	return &YOSValidator{cl: cl}, nil
}

func (r *YOSValidator) ValidateCreation(ctx context.Context, log logr.Logger, obj runtime.Object) error {
	casted := obj.(*v1.YandexObjectStorage)
	log.Info("validate create", "name", util.NamespacedName(casted))

	var key sakey.StaticAccessKey
	if err := r.cl.Get(
		ctx,
		client.ObjectKey{
			Name:      casted.Spec.SAKeyName,
			Namespace: casted.Namespace,
		},
		&key,
	); err != nil {
		if errors.IsNotFound(err) {
			return webhook.NewValidationErrorf(
				"static access key \"%s\" not found in the %s namespace", casted.Spec.SAKeyName, casted.Namespace,
			)
		}
		return fmt.Errorf("unable to get specified static access key: %w", err)
	}

	return nil
}

func (r *YOSValidator) ValidateUpdate(_ context.Context, log logr.Logger, current, old runtime.Object) error {
	castedOld, castedCurrent := old.(*v1.YandexObjectStorage), current.(*v1.YandexObjectStorage)
	log.Info("validate update", "name", util.NamespacedName(castedCurrent))

	if castedOld.Spec.Name != castedCurrent.Spec.Name {
		return webhook.NewValidationErrorf(
			"name of YandexObjectStorage must be immutable, was changed from %s to %s",
			castedOld.Spec.Name,
			castedCurrent.Spec.Name,
		)
	}

	return nil
}

func (r *YOSValidator) ValidateDeletion(ctx context.Context, log logr.Logger, obj runtime.Object) error {
	casted := obj.(*v1.YandexObjectStorage)
	log.Info("validate delete", "name", util.NamespacedName(casted))

	cred, err := awsutils.CredentialsFromStaticAccessKey(ctx, casted.Namespace, casted.Spec.SAKeyName, r.cl)
	if err != nil {
		return fmt.Errorf("unable to retrieve credentials: %w", err)
	}
	sdk, err := yosutils.NewS3Client(ctx, cred)
	if err != nil {
		return fmt.Errorf("unable to build s3 sdk: %w", err)
	}

	cnt := int64(1)
	resp, err := sdk.ListObjects(
		&s3.ListObjectsInput{
			Bucket:  &casted.Spec.Name,
			MaxKeys: &cnt,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to list objects in s3: %w", err)
	}

	if len(resp.Contents) != 0 {
		return webhook.NewValidationErrorf("cannot delete non-empty bucket %s", casted.Spec.Name)
	}

	return nil
}
