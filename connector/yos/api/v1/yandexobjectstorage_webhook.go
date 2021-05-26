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

func (r *YandexObjectStorage) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-connectors-cloud-yandex-com-v1-yandexobjectstorage,mutating=true,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexobjectstorages,verbs=create;update,versions=v1,name=myandexobjectstorage.yandex.com,admissionReviewVersions=v1

var _ webhook.Defaulter = &YandexObjectStorage{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *YandexObjectStorage) Default() {
	ycrlog.Info("default", "name", r.Name)

	// TODO (covariance): fill in your defaulting logic.
}

// TODO (covariance): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexobjectstorage,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexobjectstorages,verbs=create;update;delete,versions=v1,name=vyandexobjectstorage.yandex.com,admissionReviewVersions=v1

var _ webhook.Validator = &YandexObjectStorage{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *YandexObjectStorage) ValidateCreate() error {
	ycrlog.Info("validate create", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *YandexObjectStorage) ValidateUpdate(old runtime.Object) error {
	ycrlog.Info("validate update", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *YandexObjectStorage) ValidateDelete() error {
	ycrlog.Info("validate delete", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object deletion.
	return nil
}
