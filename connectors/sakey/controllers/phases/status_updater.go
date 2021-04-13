// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
	sakeyutils "k8s-connectors/connectors/sakey/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusUpdater struct {
	Sdk    *ycsdk.SDK
	Client *client.Client
}

func (r *StatusUpdater) IsUpdated(_ context.Context, _ *connectorsv1.StaticAccessKey) (bool, error) {
	// In every reconciliation we need to update
	// status. Therefore, this updater is never
	// marked as updated.
	return false, nil
}

func (r *StatusUpdater) Update(ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey) error {
	// We must not forget that field SecretName is
	// managed by another phase and therefore only
	// thing we do is update key cloud id.

	res, err := sakeyutils.GetStaticAccessKey(ctx, object, r.Sdk)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("resource not found in k8s")
	}

	// Do not mess this fields up, KeyId in a cloud is
	// another entity.
	object.Status.KeyID = res.Id

	if err := (*r.Client).Update(ctx, object); err != nil {
		return err
	}

	log.Info("status successfully updated")
	return nil
}

func (r *StatusUpdater) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.StaticAccessKey) error {
	return nil
}
