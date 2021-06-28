// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type mutatingHandler struct {
	object  runtime.Object
	decoder *admission.Decoder
	log     logr.Logger
	mutator Mutator
}

func NewMutatingHandler(m Mutator) admission.Handler {
	return &mutatingHandler{
		log:     logr.Discard(),
		mutator: m,
	}
}

func (r *mutatingHandler) InjectObject(obj runtime.Object) error {
	r.object = obj
	return nil
}

func (r *mutatingHandler) InjectDecoder(decoder *admission.Decoder) error {
	r.decoder = decoder
	return nil
}

func (r *mutatingHandler) InjectLogger(log logr.Logger) error {
	r.log = log
	return nil
}

func (r *mutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response { //nolint:gocritic
	// GoCritic warns about `hugeParam` req, but it is an interface that we are obliged to follow
	obj := r.object.DeepCopyObject()
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	res, err := r.mutator.Mutate(ctx, r.log, obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	marshaled, err := json.Marshal(res)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}
