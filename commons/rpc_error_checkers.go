// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package commons

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CheckRPCErrorCanceled(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.Canceled
}
func CheckRPCErrorUnknown(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.Unknown
}
func CheckRPCErrorInvalidArgument(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.InvalidArgument
}
func CheckRPCErrorDeadlineExceeded(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.DeadlineExceeded
}
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
func CheckRPCErrorPermissionDenied(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.PermissionDenied
}
func CheckRPCErrorResourceExhausted(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.ResourceExhausted
}
func CheckRPCErrorFailedPrecondition(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.FailedPrecondition
}
func CheckRPCErrorAborted(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.Aborted
}
func CheckRPCErrorOutOfRange(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.OutOfRange
}
func CheckRPCErrorUnimplemented(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.Unimplemented
}
func CheckRPCErrorInternal(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.Internal
}
func CheckRPCErrorUnavailable(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.Unavailable
}
func CheckRPCErrorDataLoss(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.DataLoss
}
func CheckRPCErrorUnauthenticated(err error) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return s.Code() == codes.Unauthenticated
}

/*
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
