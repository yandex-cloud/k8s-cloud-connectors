// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"strconv"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/ymq/api/v1"
	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
)

const (
	FifoQueue                     = "FifoQueue"
	ContentBasedDeduplication     = "ContentBasedDeduplication"
	DelaySeconds                  = "DelaySeconds"
	MaximumMessageSize            = "MaximumMessageSize"
	MessageRetentionPeriod        = "MessageRetentionPeriod"
	ReceiveMessageWaitTimeSeconds = "ReceiveMessageWaitTimeSeconds"
	VisibilityTimeout             = "VisibilityTimeout"
)

func AttributesFromSpec(spec *connectorsv1.YandexMessageQueueSpec) map[string]*string {
	attributes := map[string]*string{
		"DelaySeconds":                  util.IntToStringPtr(spec.DelaySeconds),
		"MaximumMessageSize":            util.IntToStringPtr(spec.MaximumMessageSize),
		"MessageRetentionPeriod":        util.IntToStringPtr(spec.MessageRetentionPeriod),
		"ReceiveMessageWaitTimeSeconds": util.IntToStringPtr(spec.ReceiveMessageWaitTimeSeconds),
		"VisibilityTimeout":             util.IntToStringPtr(spec.VisibilityTimeout),
	}

	if spec.FifoQueue {
		attributes["FifoQueue"] = util.StringPtr("true")
		attributes["ContentBasedDeduplication"] = util.StringPtr(strconv.FormatBool(spec.ContentBasedDeduplication))
	}
	return attributes
}
