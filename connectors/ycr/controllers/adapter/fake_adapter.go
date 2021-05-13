// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return nil, status.Errorf(codes.NotFound, "registry not found: "+registryID)
	}
	return r.Storage[registryID], nil
}

func (r *FakeYandexContainerRegistryAdapter) List(_ context.Context, folderID string) ([]*containerregistry.Registry, error) {
	var result []*containerregistry.Registry
	for _, registry := range r.Storage {
		if registry.FolderId == folderID {
			result = append(result, registry)
		}
	}
	return result, nil
}

func (r *FakeYandexContainerRegistryAdapter) Update(_ context.Context, request *containerregistry.UpdateRegistryRequest) error {
	if _, ok := r.Storage[request.RegistryId]; !ok {
		return status.Errorf(codes.NotFound, "registry not found: "+request.RegistryId)
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

func (r *FakeYandexContainerRegistryAdapter) Delete(_ context.Context, registryId string) error {
	if _, ok := r.Storage[registryId]; !ok {
		return status.Errorf(codes.NotFound, "registry not found: "+registryId)
	}
	delete(r.Storage, registryId)
	return nil
}
