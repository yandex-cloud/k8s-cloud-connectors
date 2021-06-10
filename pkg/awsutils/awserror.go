// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package awsutils

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func CheckSQSDoesNotExist(err error) bool {
	s, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	return s.Code() == sqs.ErrCodeQueueDoesNotExist
}

func CheckS3DoesNotExist(err error) bool {
	s, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	return s.Code() == s3.ErrCodeNoSuchBucket
}
