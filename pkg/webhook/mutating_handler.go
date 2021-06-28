// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type mutatingHandler struct {
	object  runtime.Object
	decoder *admission.Decoder
	mutator func(old *runtime.Object) (runtime.Object, error)
}

func (r mutatingHandler) InjectObject(obj runtime.Object) error {
	r.object = obj
	return nil
}

func (r mutatingHandler) InjectDecoder(decoder *admission.Decoder) error {
	r.decoder = decoder
	return nil
}

func (r mutatingHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	obj := r.object.DeepCopyObject()
	if err := r.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	res, err := r.mutator(&obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	marshalled, err := json.Marshal(res)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}