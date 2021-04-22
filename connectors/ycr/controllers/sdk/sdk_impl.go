// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package sdk

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/pkg/config"
	"k8s-connectors/pkg/errors"
)

type YandexContainerRegistrySDKImpl struct {
	sdk *ycsdk.SDK
}

func NewYandexContainerRegistrySDKImpl() (YandexContainerRegistrySDK, error) {
	sdk, err := ycsdk.Build(context.Background(), ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})

	if err != nil {
		return nil, err
	}
	return YandexContainerRegistrySDKImpl{
		sdk: sdk,
	}, nil
}

func checkRegistryMatchWithYcr(ycr *containerregistry.Registry, registry *connectorsv1.YandexContainerRegistry) bool {
	cluster, ok1 := ycr.Labels[config.CloudClusterLabel]
	name, ok2 := ycr.Labels[config.CloudNameLabel]
	return ok1 && ok2 && cluster == registry.ClusterName && name == registry.Name
}

func (r YandexContainerRegistrySDKImpl) Create(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Create(ctx, &containerregistry.CreateRegistryRequest{
		FolderId: registry.Spec.FolderId,
		Name:     registry.Spec.Name,
		Labels: map[string]string{
			config.CloudClusterLabel: registry.ClusterName,
			config.CloudNameLabel:    registry.Name,
		},
	}))

	if err != nil {
		// This case is quite strange, but we cannot do anything about it,
		// so we just ignore it.
		if errors.CheckRPCErrorAlreadyExists(err) {
			// TODO (covariance) is it considered error or not?
			log.Info("registry already exists")
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

	return nil
}

func (r YandexContainerRegistrySDKImpl) Read(ctx context.Context, _ logr.Logger, registry *connectorsv1.YandexContainerRegistry) (*containerregistry.Registry, error) {
	// If id is written in the status, we need to check
	// whether it exists in the cloud
	if registry.Status.Id != "" {
		ycr, err := r.sdk.ContainerRegistry().Registry().Get(ctx, &containerregistry.GetRegistryRequest{
			RegistryId: registry.Status.Id,
		})
		if err != nil {
			// If registry was not found then it does not exist,
			// but this error is not fatal, just a mismatch between
			// out status and real world state.
			if errors.CheckRPCErrorNotFound(err) {
				return nil, nil
			}
			// Otherwise, it is fatal
			return nil, fmt.Errorf("cannot get registry from cloud: %v", err)
		}

		// If labels do match with our object, then we have found it
		if checkRegistryMatchWithYcr(ycr, registry) {
			return ycr, nil
		}

		// Otherwise registry is not found, but that is ok:
		// we will try to list resources and find the one we need.
	}

	// TODO (covariance) pagination
	// Otherwise, we try to match cluster name and meta name
	// with registries in the cloud
	list, err := r.sdk.ContainerRegistry().Registry().List(ctx, &containerregistry.ListRegistriesRequest{
		FolderId: registry.Spec.FolderId,
	})
	if err != nil {
		// This error is fatal
		return nil, fmt.Errorf("cannot list registries in folder: %v", err)
	}

	for _, ycr := range list.Registries {
		// If labels do match with our object, then we have found it
		if checkRegistryMatchWithYcr(ycr, registry) {
			return ycr, nil
		}
	}

	// Nothing found, no such registry
	return nil, nil
}

func (r YandexContainerRegistrySDKImpl) Update(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	res, err := r.Read(ctx, log, registry)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("cannot update registry: %v", registry)
	}

	op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Update(ctx, &containerregistry.UpdateRegistryRequest{
		RegistryId: res.Id,
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

	return nil
}

func (r YandexContainerRegistrySDKImpl) Delete(ctx context.Context, log logr.Logger, registry *connectorsv1.YandexContainerRegistry) error {
	ycr, err := r.Read(ctx, log, registry)
	if err != nil {
		return err
	}

	if ycr != nil {
		op, err := r.sdk.WrapOperation(r.sdk.ContainerRegistry().Registry().Delete(ctx, &containerregistry.DeleteRegistryRequest{
			RegistryId: ycr.Id,
		}))
		if err != nil {
			// Not found error is already handled by getRegistryId
			return fmt.Errorf("error while deleting registry: %v", err)
		}
		if err := op.Wait(ctx); err != nil {
			return fmt.Errorf("error while deleting registry: %v", err)
		}
		log.Info("registry deleted successfully")
		return nil
	}
	// It is assumed that id is the actual id of the object since
	// its lifecycle must be fully managed by connector.
	// id being empty means that it was deleted externally,
	// thus finalization is considered complete.
	log.Info("registry deleted externally")
	return nil
}
