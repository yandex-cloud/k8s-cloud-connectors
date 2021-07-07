// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package awsutils

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func checkAWSErrorByCode(err error, code string) bool {
	var s awserr.Error
	ok := errors.As(err, &s)
	if !ok {
		return false
	}
	return s.Code() == code
}

func CheckSQSDoesNotExist(err error) bool {
	return checkAWSErrorByCode(err, sqs.ErrCodeQueueDoesNotExist)
}

func CheckS3DoesNotExist(err error) bool {
	return checkAWSErrorByCode(err, s3.ErrCodeNoSuchBucket)
}

func CheckS3AlreadyOwnedByYou(err error) bool {
	return checkAWSErrorByCode(err, s3.ErrCodeBucketAlreadyOwnedByYou)
}