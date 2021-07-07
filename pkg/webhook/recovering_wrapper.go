// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"fmt"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type RecoveringWrapper struct {
	inner admission.Handler
}

func (r *RecoveringWrapper) wrapHandler(
	ctx context.Context,
	req *admission.Request,
) (resp admission.Response, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	resp = r.inner.Handle(ctx, *req)
	return
}

func (r RecoveringWrapper) Handle(ctx context.Context, req admission.Request) admission.Response { //nolint:gocritic
	// GoCritic warns about `hugeParam` req, but it is an interface that we are obliged to follow
	resp, err := r.wrapHandler(ctx, &req)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return resp
}
