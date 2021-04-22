// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/sdk"
)

type SpecMatcher struct {
	Sdk sdk.YandexContainerRegistrySDK
}

func (r *SpecMatcher) IsUpdated(ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry) (bool, error) {
	res, err := r.Sdk.Read(ctx, log, object)
	if err != nil {
		return false, err
	}
	if res == nil {
		return false, fmt.Errorf("resource not found in cloud: %v", object)
	}

	// Here we will check immutable fields
	if object.Spec.FolderId != "" && res.FolderId != object.Spec.FolderId {
		return false, fmt.Errorf("FolderId changed, invalid state for object")
	}
	return res.Name == object.Spec.Name, nil
}

func (r *SpecMatcher) Update(ctx context.Context, log logr.Logger, object *connectorsv1.YandexContainerRegistry) error {
	if err := r.Sdk.Update(ctx, log, object); err != nil {
		return err
	}
	log.Info("object spec matched with cloud")
	return nil
}

func (r *SpecMatcher) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.YandexContainerRegistry) error {
	return nil
}
