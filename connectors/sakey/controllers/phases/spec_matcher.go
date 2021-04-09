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

func (r *SpecMatcher) Update(_ context.Context, _ logr.Logger, _ *connectorsv1.StaticAccessKey) error {
	// If update is necessary, then user has changed
	// the service account id. It must be immutable,
	// thus we will just throw an error.
	return fmt.Errorf("ServiceAccountId was changed, but must be immutable")
}

func (r *SpecMatcher) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.StaticAccessKey) error {
	return nil
}
