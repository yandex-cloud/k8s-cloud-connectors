// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"
	"strings"

	sakey "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/api/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/webhook"
)

// +kubebuilder:webhook:path=/validate-connectors-cloud-yandex-com-v1-yandexmessagequeue,mutating=false,failurePolicy=fail,sideEffects=None,groups=connectors.cloud.yandex.com,resources=yandexmessagequeues,verbs=create;update;delete,versions=v1,name=vyandexmessagequeue.yandex.com,admissionReviewVersions=v1

type YMQValidator struct {
	cl client.Client
}

func NewYMQValidator(cl client.Client) webhook.Validator {
	return &YMQValidator{cl: cl}
}

func (r *YMQValidator) ValidateCreation(ctx context.Context, log logr.Logger, obj runtime.Object) error {
	casted := obj.(*v1.YandexMessageQueue)
	log.Info("validate create", "name", util.NamespacedName(casted))

	if casted.Spec.FifoQueue && !strings.HasSuffix(casted.Spec.Name, ".fifo") {
		return webhook.NewValidationErrorf(
			"name of FIFO queue must end with \".fifo\", currently is: %s",
			casted.Spec.Name,
		)
	}

	if !casted.Spec.FifoQueue && strings.HasSuffix(casted.Spec.Name, ".fifo") {
		return webhook.NewValidationError(
			fmt.Errorf("name of non-FIFO queue must NOT end with \".fifo\", currently is: %s", casted.Spec.Name),
		)
	}

	if casted.Spec.ContentBasedDeduplication && !casted.Spec.FifoQueue {
		return webhook.NewValidationError(
			fmt.Errorf("content based deduplication is available only for FIFO queue"),
		)
	}

	var key sakey.StaticAccessKey
	if err := r.cl.Get(
		ctx,
		client.ObjectKey{
			Name:      casted.Spec.SAKeyName,
			Namespace: casted.Namespace,
		},
		&key,
	); err != nil {
		if errors.IsNotFound(err) {
			return webhook.NewValidationErrorf(
				"static access key \"%s\" not found in the %s namespace", casted.Spec.SAKeyName, casted.Namespace,
			)
		}
		return fmt.Errorf("unable to get specified static access key: %w", err)
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
