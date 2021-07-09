// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
)

type Mutator interface {
	Mutate(ctx context.Context, log logr.Logger, obj runtime.Object) (runtime.Object, error)
}

type ValidationError struct{ err error }

func (v *ValidationError) Error() string { return v.err.Error() }

func (v *ValidationError) Unwrap() error {
	return v.err
}

func (v *ValidationError) Is(err error) bool {
	_, ok := err.(*ValidationError) //nolint:errorlint
	return ok
}

func NewValidationError(inner error) *ValidationError {
	return &ValidationError{inner}
}

func NewValidationErrorf(format string, args ...interface{}) *ValidationError {
	return NewValidationError(fmt.Errorf(format, args...))
}

type Validator interface {
	ValidateCreation(ctx context.Context, log logr.Logger, obj runtime.Object) error
	ValidateUpdate(ctx context.Context, log logr.Logger, current, old runtime.Object) error
	ValidateDeletion(ctx context.Context, log logr.Logger, obj runtime.Object) error
}
