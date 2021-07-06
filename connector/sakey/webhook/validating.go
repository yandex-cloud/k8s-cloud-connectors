// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s-connectors/connector/sakey/api/v1"
	"k8s-connectors/pkg/util"
	"k8s-connectors/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-staticaccesskey,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=create;update;delete,versions=v1,name=vstaticaccesskey.yandex.com,admissionReviewVersions=v1

type SAKeyValidator struct{}

func (r *SAKeyValidator) ValidateCreation(_ context.Context, log logr.Logger, obj runtime.Object) error {
	log.Info("validate create", "name", util.NamespacedName(obj.(*v1.StaticAccessKey)))
	return nil
}

func (r *SAKeyValidator) ValidateUpdate(_ context.Context, log logr.Logger, current, old runtime.Object) error {
	castedOld, castedCurrent := old.(*v1.StaticAccessKey), current.(*v1.StaticAccessKey)

	log.Info("validate update", "name", util.NamespacedName(castedCurrent))

	if castedCurrent.Spec.ServiceAccountID != castedOld.Spec.ServiceAccountID {
		return webhook.NewValidationError(
			fmt.Errorf(
				"binded service account must be immutable, was changed from %s to %s",
				castedOld.Spec.ServiceAccountID,
				castedCurrent.Spec.ServiceAccountID,
			),
		)
	}

	return nil
}

func (r *SAKeyValidator) ValidateDeletion(_ context.Context, log logr.Logger, obj runtime.Object) error {
	log.Info("validate delete", "name", util.NamespacedName(obj.(*v1.StaticAccessKey)))
	return nil
}
