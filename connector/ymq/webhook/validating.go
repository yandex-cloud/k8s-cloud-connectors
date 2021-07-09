// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s-connectors/connector/ymq/api/v1"
	"k8s-connectors/pkg/util"
	"k8s-connectors/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexmessagequeue,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexmessagequeues,verbs=create;update;delete,versions=v1,name=vyandexmessagequeue.yandex.com,admissionReviewVersions=v1

type YMQValidator struct{}

func (r *YMQValidator) ValidateCreation(_ context.Context, log logr.Logger, obj runtime.Object) error {
	castedObj := obj.(*v1.YandexMessageQueue)
	log.Info("validate create", "name", util.NamespacedName(castedObj))

	if castedObj.Spec.FifoQueue && !strings.HasSuffix(castedObj.Spec.Name, ".fifo") {
		return webhook.NewValidationErrorf(
			"name of FIFO queue must end with \".fifo\", currently is: %s",
			castedObj.Spec.Name,
		)
	}

	if !castedObj.Spec.FifoQueue && strings.HasSuffix(castedObj.Spec.Name, ".fifo") {
		return webhook.NewValidationError(
			fmt.Errorf("name of non-FIFO queue must NOT end with \".fifo\", currently is: %s", castedObj.Spec.Name),
		)
	}

	if castedObj.Spec.ContentBasedDeduplication && !castedObj.Spec.FifoQueue {
		return webhook.NewValidationError(
			fmt.Errorf("content based deduplication is available only for FIFO queue"),
		)
	}

	return nil
}

func (r *YMQValidator) ValidateUpdate(_ context.Context, log logr.Logger, current, old runtime.Object) error {
	castedOld, castedCurrent := old.(*v1.YandexMessageQueue), current.(*v1.YandexMessageQueue)
	log.Info("validate update", "name", util.NamespacedName(castedCurrent))

	if castedCurrent.Spec.Name != castedOld.Spec.Name {
		return webhook.NewValidationError(
			fmt.Errorf(
				"name of YandexMessageQueue must be immutable, was changed from %s to %s",
				castedOld.Spec.Name,
				castedCurrent.Spec.Name,
			),
		)
	}

	if castedCurrent.Spec.FifoQueue != castedOld.Spec.FifoQueue {
		return webhook.NewValidationError(
			fmt.Errorf(
				"FIFO flag of YandexMessageQueue must be immutable, was changed from %t to %t",
				castedOld.Spec.FifoQueue,
				castedCurrent.Spec.FifoQueue,
			),
		)
	}

	return nil
}

func (r *YMQValidator) ValidateDeletion(_ context.Context, log logr.Logger, obj runtime.Object) error {
	log.Info("validate delete", "name", util.NamespacedName(obj.(*v1.YandexMessageQueue)))
	return nil
}
