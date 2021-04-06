// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrutils "k8s-connectors/connectors/ycr/pkg/utils"
)

type SpecMatcher struct {
	Sdk *ycsdk.SDK
}

func (r *SpecMatcher) IsUpdated(ctx context.Context, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	ycr, err := ycrutils.GetRegistry(ctx, registry, r.Sdk)
	if err != nil {
		return false, fmt.Errorf("unable to get registry: %v", err)
	}
	if ycr == nil {
		return false, fmt.Errorf("registry %s not found in folder %s", registry.Spec.Name, registry.Spec.FolderId)
	}

	// Here we will check immutable fields
	if registry.Spec.FolderId != "" && ycr.FolderId != registry.Spec.FolderId {
		return false, fmt.Errorf("FolderId changed, invalid state for registry")
	}
	return ycr.Name == registry.Spec.Name, nil
}

func (r *SpecMatcher) Update(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	ycr, err := ycrutils.GetRegistry(ctx, registry, r.Sdk)
	if err != nil {
		return fmt.Errorf("unable to get registry: %v", err)
	}
	if ycr == nil {
		return fmt.Errorf("registry %s not found in folder %s", registry.Spec.Name, registry.Spec.FolderId)
	}

	op, err := r.Sdk.WrapOperation(r.Sdk.ContainerRegistry().Registry().Update(ctx, &containerregistry.UpdateRegistryRequest{
		RegistryId: ycr.Id,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"name"},
		},
		Name: registry.Spec.Name,
	}))

	if err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("can't update registry in cloud: %v", err)
	}

	log.Info("registry spec matched with cloud")
	return nil
}
