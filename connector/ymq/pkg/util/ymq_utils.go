// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"strconv"

	connectorsv1 "k8s-connectors/connector/ymq/api/v1"
)

func AttributesFromSpec(spec *connectorsv1.YandexMessageQueueSpec) map[string]*string {
	delaySeconds := strconv.Itoa(spec.DelaySeconds)
	maximumMessageSize := strconv.Itoa(spec.MaximumMessageSize)
	messageRetentionPeriod := strconv.Itoa(spec.MessageRetentionPeriod)
	receiveMessageWaitTimeSeconds := strconv.Itoa(spec.ReceiveMessageWaitTimeSeconds)
	visibilityTimeout := strconv.Itoa(spec.VisibilityTimeout)

	attributes := map[string]*string{
		"DelaySeconds":                  &delaySeconds,
		"MaximumMessageSize":            &maximumMessageSize,
		"MessageRetentionPeriod":        &messageRetentionPeriod,
		"ReceiveMessageWaitTimeSeconds": &receiveMessageWaitTimeSeconds,
		"VisibilityTimeout":             &visibilityTimeout,
	}

	if spec.FifoQueue {
		fifoQueue := "true"
		contentBasedDeduplication := strconv.FormatBool(spec.ContentBasedDeduplication)
		attributes["FifoQueue"] = &fifoQueue
		attributes["ContentBasedDeduplication"] = &contentBasedDeduplication
	}
	return attributes
}
