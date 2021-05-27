// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var ymqlog = logf.Log.WithName("ymq-amdission")

func (r *YandexMessageQueue) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-connectors-cloud-yandex-com-v1-yandexmessagequeue,mutating=true,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexmessagequeues,verbs=create;update,versions=v1,name=myandexmessagequeue.yandex.com,admissionReviewVersions=v1

var _ webhook.Defaulter = &YandexMessageQueue{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *YandexMessageQueue) Default() {
}

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexmessagequeue,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexmessagequeues,verbs=create;update;delete,versions=v1,name=vyandexmessagequeue.yandex.com,admissionReviewVersions=v1

var _ webhook.Validator = &YandexMessageQueue{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *YandexMessageQueue) ValidateCreate() error {

	if r.Spec.FifoQueue {
		if !strings.HasSuffix(r.Spec.Name, ".fifo") {
			return fmt.Errorf("name of FIFO queue must end with \".fifo\"")
		}
	}

	if r.Spec.ContentBasedDeduplication && !r.Spec.FifoQueue {
		return fmt.Errorf("content based deduplication is available only for FIFO queue")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *YandexMessageQueue) ValidateUpdate(old runtime.Object) error {
	ymqlog.Info("validate update", "name", r.Name)

	oldCasted, ok := old.DeepCopyObject().(*YandexMessageQueue)

	if !ok {
		return fmt.Errorf("object is not of the YandexObjectStorage type")
	}

	if r.Spec.Name != oldCasted.Spec.Name {
		return fmt.Errorf(
			"name of YandexMessageQueue must be immutable, was changed from %s to %s",
			oldCasted.Spec.Name,
			r.Spec.Name,
		)
	}

	if r.Spec.FifoQueue != oldCasted.Spec.FifoQueue {
		return fmt.Errorf(
			"FIFO flag of YandexMessageQueue must be immutable, was changed from %s to %s",
			oldCasted.Spec.FifoQueue,
			r.Spec.FifoQueue,
		)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *YandexMessageQueue) ValidateDelete() error {
	return nil
}
