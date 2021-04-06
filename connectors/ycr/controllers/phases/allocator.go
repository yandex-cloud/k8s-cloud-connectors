// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrconfig "k8s-connectors/connectors/ycr/pkg"
	ycrutils "k8s-connectors/connectors/ycr/pkg/utils"
	"k8s-connectors/pkg/errors"
)

type Allocator struct {
	Sdk *ycsdk.SDK
}

func (r *Allocator) IsUpdated(ctx context.Context, registry *connectorsv1.YandexContainerRegistry) (bool, error) {
	ycr, err := ycrutils.GetRegistry(ctx, registry, r.Sdk)
	if err != nil {
		return false, fmt.Errorf("unable to get registry: %v", err)
	}

	return ycr != nil, nil
}

func (r *Allocator) Update(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	op, err := r.Sdk.WrapOperation(r.Sdk.ContainerRegistry().Registry().Create(ctx, &containerregistry.CreateRegistryRequest{
		FolderId: registry.Spec.FolderId,
		Name:     registry.Spec.Name,
		Labels: map[string]string{
			ycrconfig.RegistryCloudClusterLabel: registry.ClusterName,
			ycrconfig.RegistryCloudNameLabel:    registry.Name,
		},
	}))

	if err != nil {
		// This case is quite strange, but we cannot do anything about it,
		// so we just ignore it.
		if errors.CheckRPCErrorAlreadyExists(err) {
			// TODO (covariance) is it considered error or not?
			log.Info("resource already exists")
			return nil
		}
		return fmt.Errorf("error while creating registry: %v", err)
	}

	if err := op.Wait(ctx); err != nil {
		// According to SDK architecture, we do not actually need
		// to type check here. Every error here is really fatal.
		return fmt.Errorf("error while creating registry: %v", err)
	}

	if _, err := op.Response(); err != nil {
		// If we cannot get response from operation,
		// then it's totally not our responsibility.
		// And, by the way, fatal.
		return fmt.Errorf("error while creating registry: %v", err)
	}

	log.Info("registry allocated in cloud")
	return nil
}
