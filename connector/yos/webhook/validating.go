// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s-connectors/connector/yos/api/v1"
	"k8s-connectors/pkg/util"
	"k8s-connectors/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexobjectstorage,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexobjectstorages,verbs=create;update;delete,versions=v1,name=vyandexobjectstorage.yandex.com,admissionReviewVersions=v1

type YOSValidator struct{}

func (r *YOSValidator) ValidateCreation(ctx context.Context, log logr.Logger, obj runtime.Object) error {
	log.Info("validate create", "name", util.NamespacedName(obj.(*v1.YandexObjectStorage)))
	return nil
}

func (r *YOSValidator) ValidateUpdate(ctx context.Context, log logr.Logger, current, old runtime.Object) error {
	castedOld, castedCurrent := old.(*v1.YandexObjectStorage), current.(*v1.YandexObjectStorage)
	log.Info("validate update", "name", util.NamespacedName(castedCurrent))

	if castedOld.Spec.Name != castedCurrent.Spec.Name {
		return webhook.NewValidationError(fmt.Errorf(
			"name of YandexObjectStorage must be immutable, was changed from %s to %s",
			castedOld.Spec.Name,
			castedCurrent.Spec.Name,
		))
	}

	return nil
}

func (r *YOSValidator) ValidateDeletion(_ context.Context, log logr.Logger, obj runtime.Object) error {
	log.Info("validate delete", "name", util.NamespacedName(obj.(*v1.YandexObjectStorage)))
	return nil
}
