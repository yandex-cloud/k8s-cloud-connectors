// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var ycrlog = logf.Log.WithName("ycr-admission")

func (r *YandexContainerRegistry) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-connectors-cloud-yandex-com-v1-yandexcontainerregistry,mutating=true,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries,verbs=create;update,versions=v1,name=myandexcontainerregistry.yandex.com,admissionReviewVersions=v1

var _ webhook.Defaulter = &YandexContainerRegistry{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *YandexContainerRegistry) Default() {
}

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexcontainerregistry,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries,verbs=create;update;delete,versions=v1,name=vyandexcontainerregistry.yandex.com,admissionReviewVersions=v1

var _ webhook.Validator = &YandexContainerRegistry{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *YandexContainerRegistry) ValidateCreate() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *YandexContainerRegistry) ValidateUpdate(old runtime.Object) error {
	ycrlog.Info("validate update", "name", r.Name)

	oldCasted, ok := old.DeepCopyObject().(*YandexContainerRegistry)

	if !ok {
		return fmt.Errorf("object is not of the YandexContainerRegistry type")
	}

	if r.Spec.Name != oldCasted.Spec.Name {
		return fmt.Errorf(
			"name of YandexContainerRegistry must be immutable, was changed from %s to %s",
			oldCasted.Spec.Name,
			r.Spec.Name,
		)
	}

	if r.Spec.FolderID != oldCasted.Spec.FolderID {
		return fmt.Errorf(
			"folder id of YandexContainerRegistry must be immutable, was changed from %s to %s",
			oldCasted.Spec.FolderID,
			r.Spec.FolderID,
		)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *YandexContainerRegistry) ValidateDelete() error {
	return nil
}
