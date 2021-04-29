// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/adapter"
	ycrutils "k8s-connectors/connectors/ycr/pkg/util"
)

type SpecMatcher struct {
	Sdk adapter.YandexContainerRegistryAdapter
}

func (r *SpecMatcher) IsUpdated(ctx context.Context, _ logr.Logger, object *connectorsv1.YandexContainerRegistry) (bool, error) {
	res, err := ycrutils.GetRegistry(ctx, object.Status.Id, object.Spec.FolderId, object.ObjectMeta.Name, object.ObjectMeta.ClusterName, r.Sdk)
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
	ycr, err := ycrutils.GetRegistry(ctx, object.Status.Id, object.Spec.FolderId, object.ObjectMeta.Name, object.ObjectMeta.ClusterName, r.Sdk)
	if err != nil {
		return err
	}
	if ycr == nil {
		return fmt.Errorf("object does not exist in the cloud")
	}

	if err := r.Sdk.Update(ctx, &containerregistry.UpdateRegistryRequest{
		RegistryId: ycr.Id,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		Name:       object.Spec.Name,
	}); err != nil {
		return err
	}
	log.Info("object spec matched with cloud")
	return nil
}

func (r *SpecMatcher) Cleanup(_ context.Context, _ logr.Logger, _ *connectorsv1.YandexContainerRegistry) error {
	return nil
}
