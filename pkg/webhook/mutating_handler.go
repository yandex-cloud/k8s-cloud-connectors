// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type mutatingHandler struct {
	object  runtime.Object
	decoder *admission.Decoder
	log     logr.Logger
	mutator Mutator
}

func RegisterMutatingHandler(mgr manager.Manager, exemplar runtime.Object, m Mutator) error {
	decoder, err := admission.NewDecoder(mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("unable to create decoder for scheme: %w", err)
	}

	if err := RegisterForManager(
		mgr,
		exemplar,
		&mutatingHandler{
			object:  exemplar,
			decoder: decoder,
			log:     mgr.GetLogger(),
			mutator: m,
		},
		"mutate",
	); err != nil {
		return err
	}

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
