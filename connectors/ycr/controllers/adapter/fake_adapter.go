// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"strconv"
)

type FakeYandexContainerRegistryAdapter struct {
	Storage map[string]*containerregistry.Registry
	FreeId  int
}

func NewFakeYandexContainerRegistryAdapter() FakeYandexContainerRegistryAdapter {
	return FakeYandexContainerRegistryAdapter{
		Storage: map[string]*containerregistry.Registry{},
		FreeId:  0,
	}
}

func (r *FakeYandexContainerRegistryAdapter) Create(_ context.Context, request *containerregistry.CreateRegistryRequest) (*containerregistry.Registry, error) {
	// TODO (covariance) Remember that this is not intended behaviour and in future YCR must be checked for name uniqueness
	registry := containerregistry.Registry{
		Id:        strconv.Itoa(r.FreeId),
		FolderId:  request.FolderId,
		Name:      request.Name,
		CreatedAt: ptypes.TimestampNow(),
		Labels:    request.Labels,
	}
	r.Storage[strconv.Itoa(r.FreeId)] = &registry
	r.FreeId++
	return &registry, nil
}

func (r *FakeYandexContainerRegistryAdapter) Read(_ context.Context, registryID string) (*containerregistry.Registry, error) {
	if _, ok := r.Storage[registryID]; !ok {
		return nil, fmt.Errorf("registry not found")
	}
	return r.Storage[registryID], nil
}

func (r *FakeYandexContainerRegistryAdapter) List(_ context.Context, folderID string) ([]*containerregistry.Registry, error) {
	result := []*containerregistry.Registry{}
	for _, registry := range r.Storage {
		if registry.FolderId == folderID {
			result = append(result, registry)
		}
	}
	return result, nil
}

func (r *FakeYandexContainerRegistryAdapter) Update(_ context.Context, request *containerregistry.UpdateRegistryRequest) error {
	if _, ok := r.Storage[request.RegistryId]; !ok {
		return fmt.Errorf("registry not found")
	}
	for _, path := range request.UpdateMask.Paths {
		if path == "name" {
			r.Storage[request.RegistryId].Name = request.Name
		}
		if path == "labels" {
			r.Storage[request.RegistryId].Labels = request.Labels
		}
	}
	return nil
}

func (r *FakeYandexContainerRegistryAdapter) Delete(_ context.Context, request *containerregistry.DeleteRegistryRequest) error {
	if _, ok := r.Storage[request.RegistryId]; !ok {
		return fmt.Errorf("registry not found")
	}
	delete(r.Storage, request.RegistryId)
	return nil
}
