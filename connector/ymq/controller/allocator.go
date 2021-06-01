// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/ymq/api/v1"
	ymqutils "k8s-connectors/connector/ymq/pkg/util"
)

func (r *yandexMessageQueueReconciler) allocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexMessageQueue,
) error {
	log.V(1).Info("started")

	key, secret, err := ymqutils.KeyAndSecretFromStaticAccessKey(ctx, object, r.Client)
	if err != nil {
		return fmt.Errorf("unable to retrieve key and secret: %v", err)
	}

	lst, err := r.adapter.List(ctx, key, secret)
	if err != nil {
		return fmt.Errorf("unable to list resources: %v", err)
	}
	for _, queue := range lst {
		if *queue == object.Status.QueueURL {
			return nil
		}
	}

	delaySeconds := strconv.Itoa(object.Spec.DelaySeconds)
	maximumMessageSize := strconv.Itoa(object.Spec.MaximumMessageSize)
	messageRetentionPeriod := strconv.Itoa(object.Spec.MessageRetentionPeriod)
	receiveMessageWaitTimeSeconds := strconv.Itoa(object.Spec.ReceiveMessageWaitTimeSeconds)
	visibilityTimeout := strconv.Itoa(object.Spec.VisibilityTimeout)

	attributes := map[string]*string{
		"DelaySeconds":                  &delaySeconds,
		"MaximumMessageSize":            &maximumMessageSize,
		"MessageRetentionPeriod":        &messageRetentionPeriod,
		"ReceiveMessageWaitTimeSeconds": &receiveMessageWaitTimeSeconds,
		"VisibilityTimeout":             &visibilityTimeout,
	}

	if object.Spec.FifoQueue {
		fifoQueue := "true"
		contentBasedDeduplication := strconv.FormatBool(object.Spec.ContentBasedDeduplication)
		attributes["FifoQueue"] = &fifoQueue
		attributes["ContentBasedDeduplication"] = &contentBasedDeduplication
	}

	res, err := r.adapter.Create(ctx, key, secret, attributes, object.Spec.Name)
	if err != nil {
		return fmt.Errorf("ubable to create resource: %v", err)
	}

	object.Status.QueueURL = res
	if err := r.Client.Status().Update(ctx, object); err != nil {
		return fmt.Errorf("unable to update object status: %v", err)
	}

	log.Info("successful")
	return nil
}

func (r *yandexMessageQueueReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, object *connectorsv1.YandexMessageQueue,
) error {
	log.V(1).Info("started")

	key, secret, err := ymqutils.KeyAndSecretFromStaticAccessKey(ctx, object, r.Client)
	if err != nil {
		return fmt.Errorf("unable to retrieve key and secret: %v", err)
	}

	err = r.adapter.Delete(ctx, key, secret, object.Status.QueueURL)
	if err != nil {
		return fmt.Errorf("unable to delete resource: %v", err)
	}

	log.Info("successful")
	return nil
}
