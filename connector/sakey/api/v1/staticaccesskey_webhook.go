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
var sakeylog = logf.Log.WithName("sakey-admission")

func (r *StaticAccessKey) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-connectors-cloud-yandex-com-v1-staticaccesskey,mutating=true,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=create;update,versions=v1,name=mstaticaccesskey.yandex.com,admissionReviewVersions=v1

var _ webhook.Defaulter = &StaticAccessKey{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *StaticAccessKey) Default() {
}

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-staticaccesskey,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=staticaccesskeys,verbs=create;update;delete,versions=v1,name=vstaticaccesskey.yandex.com,admissionReviewVersions=v1

var _ webhook.Validator = &StaticAccessKey{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *StaticAccessKey) ValidateCreate() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *StaticAccessKey) ValidateUpdate(old runtime.Object) error {
	sakeylog.Info("validate update", "name", r.Name)

	oldCasted, ok := old.DeepCopyObject().(*StaticAccessKey)

	if !ok {
		return fmt.Errorf("object is not of the StaticAccessKey type")
	}

	if r.Spec.ServiceAccountID != oldCasted.Spec.ServiceAccountID {
		return fmt.Errorf(
			"binded service account must be immutable, was changed from %s to %s",
			oldCasted.Spec.ServiceAccountID,
			r.Spec.ServiceAccountID,
		)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *StaticAccessKey) ValidateDelete() error {
	return nil
}
