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
	ctx context.Context, log logr.Logger, resource *connectorsv1.YandexMessageQueue,
) error {
	key, secret, err := ymqutils.KeyAndSecretFromStaticAccessKey(ctx, resource, r.Client)
	if err != nil {
		return err
	}

	lst, err := r.adapter.List(ctx, key, secret)
	if err != nil {
		return err
	}
	for _, queue := range lst {
		if *queue == resource.Status.QueueURL {
			return nil
		}
	}

	delaySeconds := strconv.Itoa(resource.Spec.DelaySeconds)
	maximumMessageSize := strconv.Itoa(resource.Spec.MaximumMessageSize)
	messageRetentionPeriod := strconv.Itoa(resource.Spec.MessageRetentionPeriod)
	receiveMessageWaitTimeSeconds := strconv.Itoa(resource.Spec.ReceiveMessageWaitTimeSeconds)
	visibilityTimeout := strconv.Itoa(resource.Spec.VisibilityTimeout)

	attributes := map[string]*string{
		"DelaySeconds":                  &delaySeconds,
		"MaximumMessageSize":            &maximumMessageSize,
		"MessageRetentionPeriod":        &messageRetentionPeriod,
		"ReceiveMessageWaitTimeSeconds": &receiveMessageWaitTimeSeconds,
		"VisibilityTimeout":             &visibilityTimeout,
	}

	if resource.Spec.FifoQueue {
		fifoQueue := "true"
		contentBasedDeduplication := strconv.FormatBool(resource.Spec.ContentBasedDeduplication)
		attributes["FifoQueue"] = &fifoQueue
		attributes["ContentBasedDeduplication"] = &contentBasedDeduplication
	}

	res, err := r.adapter.Create(ctx, key, secret, attributes, resource.Spec.Name)
	if err != nil {
		return err
	}

	resource.Status.QueueURL = res
	if err := r.Client.Status().Update(ctx, resource); err != nil {
		return fmt.Errorf("error while creating resource: %v", err)
	}

	log.Info("resource successfully allocated")
	return nil
}

func (r *yandexMessageQueueReconciler) deallocateResource(
	ctx context.Context, log logr.Logger, resource *connectorsv1.YandexMessageQueue,
) error {
	key, secret, err := ymqutils.KeyAndSecretFromStaticAccessKey(ctx, resource, r.Client)
	if err != nil {
		return err
	}

	err = r.adapter.Delete(ctx, key, secret, resource.Status.QueueURL)
	if err != nil {
		return err
	}

	log.Info("resource successfully deleted")
	return nil
}
