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
var sakeylog = logf.Log.WithName("sakey-resource")

func (r *StaticAccessKey) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-connectors-cloud-yandex-com-v1-staticaccesskey,mutating=true,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=create;update,versions=v1,name=mstaticaccesskey.yandex.com,admissionReviewVersions=v1

var _ webhook.Defaulter = &StaticAccessKey{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *StaticAccessKey) Default() {
	sakeylog.Info("default", "name", r.Name)

	// TODO (covariance): fill in your defaulting logic.
}

// TODO (covariance): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-staticaccesskey,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=create;update;delete,versions=v1,name=vstaticaccesskey.yandex.com,admissionReviewVersions=v1

var _ webhook.Validator = &StaticAccessKey{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *StaticAccessKey) ValidateCreate() error {
	sakeylog.Info("validate create", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *StaticAccessKey) ValidateUpdate(old runtime.Object) error {
	sakeylog.Info("validate update", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *StaticAccessKey) ValidateDelete() error {
	sakeylog.Info("validate delete", "name", r.Name)

	// TODO (covariance): fill in your validation logic upon object deletion.
	return nil
}
