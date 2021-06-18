// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/ymq/api/v1"
	ymqutils "k8s-connectors/connector/ymq/pkg/util"
	"k8s-connectors/pkg/awsutils"
)

func (r *yandexMessageQueueReconciler) allocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexMessageQueue, sdk *sqs.SQS,
) error {
	log.V(1).Info("started")

	lst, err := r.adapter.List(ctx, sdk)
	if err != nil {
		return fmt.Errorf("unable to list resources: %w", err)
	}
	for _, queue := range lst {
		if *queue == object.Status.QueueURL {
			return nil
		}
	}

	res, err := r.adapter.Create(ctx, sdk, ymqutils.AttributesFromSpec(&object.Spec), object.Spec.Name)
	if err != nil {
		return fmt.Errorf("ubable to create resource: %w", err)
	}

	object.Status.QueueURL = res
	if err := r.Client.Status().Update(ctx, object); err != nil {
		return fmt.Errorf("unable to update object status: %w", err)
	}

	log.Info("successful")
	return nil
}

func (r *yandexMessageQueueReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexMessageQueue, sdk *sqs.SQS,
) error {
	log.V(1).Info("started")

	err := r.adapter.Delete(ctx, sdk, object.Status.QueueURL)
	if err != nil {
		if awsutils.CheckSQSDoesNotExist(err) {
			log.Info("already deleted")
			return nil
		}
		return fmt.Errorf("unable to delete resource: %w", err)
	}

	log.Info("successful")
	return nil
}
