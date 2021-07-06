// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s-connectors/connector/ycr/api/v1"
	"k8s-connectors/pkg/util"
	"k8s-connectors/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexcontainerregistry,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries,verbs=create;update;delete,versions=v1,name=vyandexcontainerregistry.yandex.com,admissionReviewVersions=v1

type YCRValidator struct{}

func (r *YCRValidator) ValidateCreation(_ context.Context, log logr.Logger, obj runtime.Object) error {
	log.Info("validate create", "name", util.NamespacedName(obj.(*v1.YandexContainerRegistry)))
	return nil
}

func (r *YCRValidator) ValidateUpdate(_ context.Context, log logr.Logger, current, old runtime.Object) error {
	castedOld, castedCurrent := old.(*v1.YandexContainerRegistry), current.(*v1.YandexContainerRegistry)

	log.Info("validate update", "name", util.NamespacedName(castedCurrent))

	if castedCurrent.Spec.FolderID != castedOld.Spec.FolderID {
		return webhook.NewValidationError(
			fmt.Errorf(
				"folder id must be immutable, was changed from %s to %s",
				castedOld.Spec.FolderID,
				castedCurrent.Spec.FolderID,
			),
		)
	}

	return nil
}

func (r *YCRValidator) ValidateDeletion(_ context.Context, log logr.Logger, obj runtime.Object) error {
	log.Info("validate delete", "name", util.NamespacedName(obj.(*v1.YandexContainerRegistry)))
	return nil
}
