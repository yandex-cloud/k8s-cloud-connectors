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
)

func (r *yandexMessageQueueReconciler) matchSpec(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexMessageQueue, sdk *sqs.SQS,
) error {
	log.V(1).Info("started")

	attributes := ymqutils.AttributesFromSpec(&object.Spec)
	oldAttributes, err := r.adapter.GetAttributes(ctx, sdk, object.Status.QueueURL)
	if err != nil {
		return fmt.Errorf("unable to get queue attributes: %v", err)
	}

	for k, v := range attributes {
		if *oldAttributes[k] != *v {
			log.V(1).Info("arguments do not match, updating")
			if err := r.adapter.UpdateAttributes(ctx, sdk, attributes, object.Status.QueueURL); err != nil {
				return fmt.Errorf("unable to update attributes: %v", err)
			}
			log.Info("successful")
			return nil
		}
	}

	return nil
}
