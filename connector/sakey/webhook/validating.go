// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/errorhandling"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-staticaccesskey,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=create;update;delete,versions=v1,name=vstaticaccesskey.yandex.com,admissionReviewVersions=v1

type SAKeyValidator struct {
	sdk *ycsdk.SDK
}

func NewSAKeyValidator(ctx context.Context) (webhook.Validator, error) {
	sdk, err := ycsdk.Build(
		ctx,
		ycsdk.Config{
			Credentials: ycsdk.InstanceServiceAccount(),
		},
	)
	if err != nil {
		return nil, err
	}
	return &SAKeyValidator{sdk: sdk}, nil
}

func (r *SAKeyValidator) ValidateCreation(ctx context.Context, log logr.Logger, obj runtime.Object) error {
	casted := obj.(*v1.StaticAccessKey)

	log.Info("validate create", "name", util.NamespacedName(casted))

	if _, err := r.sdk.IAM().ServiceAccount().Get(
		ctx, &iam.GetServiceAccountRequest{
			ServiceAccountId: casted.Spec.ServiceAccountID,
		},
	); err != nil {
		if errorhandling.CheckRPCErrorNotFound(err) {
			return webhook.NewValidationErrorf(
				"service account cannot be found in the cloud: %s",
				casted.Spec.ServiceAccountID,
			)
		}
		return fmt.Errorf("unable to get service account: %w", err)
	}

	return nil
}

func (r *SAKeyValidator) ValidateUpdate(_ context.Context, log logr.Logger, current, old runtime.Object) error {
	castedOld, castedCurrent := old.(*v1.StaticAccessKey), current.(*v1.StaticAccessKey)

	log.Info("validate update", "name", util.NamespacedName(castedCurrent))

	if castedCurrent.Spec.ServiceAccountID != castedOld.Spec.ServiceAccountID {
		return webhook.NewValidationErrorf(
			"bound service account must be immutable, was changed from %s to %s",
			castedOld.Spec.ServiceAccountID,
			castedCurrent.Spec.ServiceAccountID,
		)
	}

	return nil
}

func (r *SAKeyValidator) ValidateDeletion(_ context.Context, log logr.Logger, obj runtime.Object) error {
	log.Info("validate delete", "name", util.NamespacedName(obj.(*v1.StaticAccessKey)))
	return nil
}
