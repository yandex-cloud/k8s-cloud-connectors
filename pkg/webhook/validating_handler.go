// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type validatingHandler struct {
	object    runtime.Object
	decoder   *admission.Decoder
	log       logr.Logger
	validator Validator
}

func RegisterValidatingHandler(mgr manager.Manager, exemplar runtime.Object, v Validator) error {
	decoder, err := admission.NewDecoder(mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("unable to create decoder for scheme: %w", err)
	}

	if err := RegisterForManager(
		mgr,
		exemplar,
		&validatingHandler{
			object:    exemplar,
			decoder:   decoder,
			log:       mgr.GetLogger(),
			validator: v,
		},
		"validate",
	); err != nil {
		return err
	}

	return nil
}

func (r *validatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response { //nolint:gocritic
	// GoCritic warns about `hugeParam` req, but it is an interface that we are obliged to follow
	obj := r.object.DeepCopyObject()

	// It is `exhaustive` because of default case, but linter seems to raise false positive error here
	switch req.Operation { //nolint:exhaustive
	case v1.Create:
		if err := r.decoder.Decode(req, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		return handleValidationError(r.validator.ValidateCreation(ctx, r.log, obj))
	case v1.Update:
		old := r.object.DeepCopyObject()
		if err := r.decoder.DecodeRaw(req.Object, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to decode current object: %w", err))
		}
		if err := r.decoder.DecodeRaw(req.OldObject, old); err != nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("unable to decode old object: %w", err))
		}
		return handleValidationError(r.validator.ValidateUpdate(ctx, r.log, obj, old))
	case v1.Delete:
		if err := r.decoder.DecodeRaw(req.OldObject, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		return handleValidationError(r.validator.ValidateDeletion(ctx, r.log, obj))
	default:
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid request operation: %s", req.Operation))
	}
}

func handleValidationError(err error) admission.Response {
	if err != nil {
		if errors.Is(err, &ValidationError{}) {
			return admission.Denied(err.Error())
		}
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.Allowed("")
}
