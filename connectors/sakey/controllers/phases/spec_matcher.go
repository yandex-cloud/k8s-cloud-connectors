// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
	sakeyutils "k8s-connectors/connectors/sakey/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SpecMatcher struct {
	Sdk    *ycsdk.SDK
	Client *client.Client
}

func (r *SpecMatcher) IsUpdated(ctx context.Context, object *connectorsv1.StaticAccessKey) (bool, error) {
	res, err := sakeyutils.GetStaticAccessKey(ctx, object, r.Sdk)
	if err != nil {
		return false, err
	}

	return res.ServiceAccountId != object.Spec.ServiceAccountID, nil
}

func (r *SpecMatcher) Update(ctx context.Context, log logr.Logger, object *connectorsv1.StaticAccessKey) error {
	// If update is necessary, then user has changed
	// the service account id. We need to delete old
	// key and secret and let it be, because next
	// reconciliation phase will allocate them anew.
	if err := sakeyutils.DeleteStaticAccessKeyAndSecret(ctx, r.Client, r.Sdk, object); err != nil {
		return err
	}

	log.Info("spec updated successfully")
	return nil
}

func (r *SpecMatcher) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.StaticAccessKey) error {
	return nil
}
