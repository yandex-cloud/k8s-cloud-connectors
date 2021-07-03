// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/api/v1"
)

func (r *staticAccessKeyReconciler) updateStatus(
	ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey, res *awscompatibility.AccessKey,
) error {
	log.V(1).Info("started")
	// We must not forget that field SecretName is
	// managed by another phase and therefore only
	// thing we do is update key cloud id.

	// Do not mess this fields up, KeyId in a cloud is
	// another entity.
	if object.Status.KeyID == res.Id {
		return nil
	}

	object.Status.KeyID = res.Id
	if err := r.Client.Update(ctx, object); err != nil {
		return fmt.Errorf("unable to update object status: %w", err)
	}

	log.Info("successful")
	return nil
}
