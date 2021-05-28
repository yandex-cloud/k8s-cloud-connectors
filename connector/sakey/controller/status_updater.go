// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	connectorsv1 "k8s-connectors/connector/sakey/api/v1"
	sakeyutils "k8s-connectors/connector/sakey/pkg/util"
)

func (r *staticAccessKeyReconciler) updateStatus(
	ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey,
) error {
	log.V(1).Info("started")
	// We must not forget that field SecretName is
	// managed by another phase and therefore only
	// thing we do is update key cloud id.

	res, err := sakeyutils.GetStaticAccessKey(
		ctx, object.Status.KeyID, object.Spec.ServiceAccountID, r.clusterID, object.Name, r.adapter,
	)
	if err != nil {
		return fmt.Errorf("unable to get resource: %v", err)
	}
	if res == nil {
		return fmt.Errorf("resource not found in the cloud")
	}

	// Do not mess this fields up, KeyId in a cloud is
	// another entity.
	object.Status.KeyID = res.Id

	if err := r.Client.Update(ctx, object); err != nil {
		return fmt.Errorf("unable to update object status: %v", err)
	}

	log.Info("successful")
	return nil
}
