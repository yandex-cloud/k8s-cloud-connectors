// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/awskey/api/v1"
	awskeyutils "k8s-connectors/connectors/awskey/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SpecMatcher struct {
	Sdk    *ycsdk.SDK
	Client *client.Client
}

func (r *SpecMatcher) IsUpdated(ctx context.Context, object *connectorsv1.AWSAccessKey) (bool, error) {
	res, err := awskeyutils.GetAWSAccessKey(ctx, object, r.Sdk)
	if err != nil {
		return false, err
	}

	return res.ServiceAccountId != object.Spec.ServiceAccountID, nil
}

func (r *SpecMatcher) Update(ctx context.Context, log logr.Logger, object *connectorsv1.AWSAccessKey) error {
	// If update is necessary, then user has changed
	// the service account id. We need to delete old
	// key and secret and let it be, because next
	// reconciliation phase will allocate them anew.
	if err := awskeyutils.DeleteAWSAccessKeyAndSecret(ctx, r.Client, r.Sdk, object); err != nil {
		return err
	}

	log.Info("spec updated successfully")
	return nil
}

func (r *SpecMatcher) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.AWSAccessKey) error {
	return nil
}
