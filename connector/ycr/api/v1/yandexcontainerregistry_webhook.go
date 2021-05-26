// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var ycrlog = logf.Log.WithName("sakey-resource")

func (r *YandexContainerRegistry) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-connectors-cloud-yandex-com-v1-yandexcontainerregistry,mutating=true,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries,verbs=create;update,versions=v1,name=myandexcontainerregistry.yandex.com,admissionReviewVersions=v1

var _ webhook.Defaulter = &YandexContainerRegistry{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *YandexContainerRegistry) Default() {
	ycrlog.Info("default", "name", r.Name)

	// TODO (covariance): fill in your defaulting logic.
}

// TODO (covariance): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexcontainerregistry,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexcontainerregistries,verbs=create;update;delete,versions=v1,name=vyandexcontainerregistry.yandex.com,admissionReviewVersions=v1

var _ webhook.Validator = &YandexContainerRegistry{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *YandexContainerRegistry) ValidateCreate() error {
	ycrlog.Info("validate create", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *YandexContainerRegistry) ValidateUpdate(old runtime.Object) error {
	ycrlog.Info("validate update", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *YandexContainerRegistry) ValidateDelete() error {
	ycrlog.Info("validate delete", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object deletion.
	return nil
}
