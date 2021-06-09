// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package errorhandling

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
All possible errors:

Canceled
Unknown
InvalidArgument
DeadlineExceeded
NotFound
AlreadyExists
PermissionDenied
ResourceExhausted
FailedPrecondition
Aborted
OutOfRange
Unimplemented
Internal
Unavailable
DataLoss
Unauthenticated
*/

func CheckRPCErrorNotFound(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.NotFound
}

func CheckRPCErrorAlreadyExists(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.AlreadyExists
}
