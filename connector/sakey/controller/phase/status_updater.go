// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connector/sakey/api/v1"
	"k8s-connectors/connector/sakey/controller/adapter"
	sakeyutils "k8s-connectors/connector/sakey/pkg/util"
)

type StatusUpdater struct {
	Sdk       adapter.StaticAccessKeyAdapter
	Client    client.Client
	ClusterID string
}

func (r *StatusUpdater) IsUpdated(_ context.Context, _ logr.Logger, _ *connectorsv1.StaticAccessKey) (bool, error) {
	// In every reconciliation we need to update
	// status. Therefore, this updater is never
	// marked as updated.
	return false, nil
}

func (r *StatusUpdater) Update(ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey) error {
	// We must not forget that field SecretName is
	// managed by another phase and therefore only
	// thing we do is update key cloud id.

	res, err := sakeyutils.GetStaticAccessKey(
		ctx, object.Status.KeyID, object.Spec.ServiceAccountID, r.ClusterID, object.Name, r.Sdk,
	)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("resource not found in k8s")
	}

	// Do not mess this fields up, KeyId in a cloud is
	// another entity.
	object.Status.KeyID = res.Id

	if err := r.Client.Update(ctx, object); err != nil {
		return err
	}

	log.Info("status successfully updated")
	return nil
}

func (r *StatusUpdater) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.StaticAccessKey) error {
	return nil
}
