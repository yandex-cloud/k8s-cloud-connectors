// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/adapter"
	"k8s-connectors/pkg/config"
	logrfake "k8s-connectors/testing/logr-fake"
	"testing"
)

func setupAllocator(t *testing.T) (context.Context, logr.Logger, adapter.YandexContainerRegistryAdapter, YandexContainerRegistryPhase) {
	ad := adapter.NewFakeYandexContainerRegistryAdapter()
	return context.Background(), logrfake.NewFakeLogger(t), &ad, &Allocator{Sdk: &ad}
}

func createResource(ctx context.Context, ad adapter.YandexContainerRegistryAdapter, t *testing.T, specName, folderId, metaName, clusterName string) *containerregistry.Registry {
	res, err := ad.Create(ctx, &containerregistry.CreateRegistryRequest{
		FolderId: folderId,
		Name:     specName,
		Labels: map[string]string{
			config.CloudClusterLabel: clusterName,
			config.CloudNameLabel:    metaName,
		},
	})
	require.NoError(t, err)
	return res
}

func createResourceFromObject(ctx context.Context, ad adapter.YandexContainerRegistryAdapter, t *testing.T, object connectorsv1.YandexContainerRegistry) *containerregistry.Registry {
	return createResource(ctx, ad, t, object.Spec.Name, object.Spec.FolderId, object.Name, object.ClusterName)
}

func TestAllocatorIsUpdated(t *testing.T) {
	t.Run("is not updated on empty cloud", func(t *testing.T) {
		// Arrange
		ctx, log, _, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("is updated on cloud with only this registry", func(t *testing.T) {
		// Arrange
		ctx, log, ad, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")
		res := createResourceFromObject(ctx, ad, t, obj)
		obj.Status.Id = res.Id

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("is not updated on cloud with other registries", func(t *testing.T) {
		// Arrange
		ctx, log, ad, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")

		_ = createResource(ctx, ad, t, "resource1", "folder", "obj1", "")
		_ = createResource(ctx, ad, t, "resource2", "other-folder", "obj2", "")

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("is updated on cloud with this and other registries", func(t *testing.T) {
		// Arrange
		ctx, log, ad, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")

		res := createResourceFromObject(ctx, ad, t, obj)
		obj.Status.Id = res.Id

		_ = createResource(ctx, ad, t, "resource1", "folder", "obj1", "")
		_ = createResource(ctx, ad, t, "resource2", "other-folder", "obj2", "")

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})
}

func TestAllocatorUpdate(t *testing.T) {
	t.Run("update on empty cloud creates resource", func(t *testing.T) {
		// Arrange
		ctx, log, _, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")

		// Act
		require.NoError(t, phase.Update(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("update on non-empty cloud creates resource", func(t *testing.T) {
		// Arrange
		ctx, log, ad, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")

		_ = createResource(ctx, ad, t, "resource1", "folder", "obj1", "")
		_ = createResource(ctx, ad, t, "resource2", "other-folder", "obj2", "")

		// Act
		require.NoError(t, phase.Update(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})
}

func TestAllocatorCleanup(t *testing.T) {
	t.Run("cleanup on cloud with resource deletes resource", func(t *testing.T) {
		// Arrange
		ctx, log, ad, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")

		require.NoError(t, phase.Update(ctx, log, &obj))

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)
		lst, err := ad.List(ctx, "folder")
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
		assert.Len(t, lst, 0)
	})

	t.Run("cleanup on cloud with this and other resources deletes this resource", func(t *testing.T) {
		// Arrange
		ctx, log, ad, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")

		otherObj1 := CreateObject("other-resource", "folder", "otherObj1", "default")
		otherObj2 := CreateObject("resource", "other-folder", "otherObj2", "default")

		require.NoError(t, phase.Update(ctx, log, &obj))
		require.NoError(t, phase.Update(ctx, log, &otherObj1))
		require.NoError(t, phase.Update(ctx, log, &otherObj2))

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)
		lst1, err := ad.List(ctx, "folder")
		require.NoError(t, err)
		lst2, err := ad.List(ctx, "other-folder")
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
		assert.Len(t, lst1, 1)
		assert.Len(t, lst2, 1)
	})

	t.Run("cleanup on cloud without resource does nothing", func(t *testing.T) {
		// Arrange
		ctx, log, ad, phase := setupAllocator(t)
		obj := CreateObject("resource", "folder", "obj", "default")

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)
		lst, err := ad.List(ctx, "folder")
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
		assert.Len(t, lst, 0)
	})
}
