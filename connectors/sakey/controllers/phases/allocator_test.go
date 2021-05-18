// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s-connectors/connectors/sakey/controllers/adapter"
	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func setupAllocator(t *testing.T) (context.Context, logr.Logger, client.Client, adapter.StaticAccessKeyAdapter, StaticAccessKeyPhase) {
	ad := adapter.NewFakeStaticAccessKeyAdapter()
	cl := k8sfake.NewFakeClient()
	return context.Background(), logrfake.NewFakeLogger(t), cl, &ad, &Allocator{
		Sdk:    &ad,
		Client: &cl,
	}
}

func TestAllocatorIsUpdated(t *testing.T) {
	t.Run("is not updated on empty cloud", func(t *testing.T) {
		// Arrange
		ctx, log, _, _, phase := setupAllocator(t)
		obj := createObject("sukhov", "obj", "default")

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("is updated on cloud with only this resource", func(t *testing.T) {
		// Arrange
		ctx, log, _, ad, phase := setupAllocator(t)
		obj := createObject("sukhov", "obj", "default")
		_, err := ad.Create(ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj.ClusterName, obj.Name))
		require.NoError(t, err)

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("is not updated on other resources", func(t *testing.T) {
		// Arrange
		ctx, log, _, ad, phase := setupAllocator(t)
		obj1 := createObject("sukhov", "obj1", "default")
		_, err := ad.Create(ctx, obj1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj1.ClusterName, obj1.Name))
		require.NoError(t, err)
		obj2 := createObject("abdullah", "obj2", "other-namespace")
		_, err = ad.Create(ctx, obj2.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj2.ClusterName, obj2.Name))
		require.NoError(t, err)

		obj3 := createObject("gulchatay", "obj3", "default")

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj3)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("is updated on this and other resources", func(t *testing.T) {
		// Arrange
		ctx, log, _, ad, phase := setupAllocator(t)
		obj1 := createObject("sukhov", "obj1", "default")
		_, err := ad.Create(ctx, obj1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj1.ClusterName, obj1.Name))
		require.NoError(t, err)
		obj2 := createObject("abdullah", "obj2", "other-namespace")
		_, err = ad.Create(ctx, obj2.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj2.ClusterName, obj2.Name))
		require.NoError(t, err)
		obj3 := createObject("gulchatay", "obj3", "default")
		_, err = ad.Create(ctx, obj3.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj3.ClusterName, obj3.Name))
		require.NoError(t, err)

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj3)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})
}

func TestAllocatorUpdate(t *testing.T) {
	t.Run("update on empty cloud creates resource", func(t *testing.T) {
		// Arrange
		ctx, log, cl, _, phase := setupAllocator(t)
		obj := createObject("sukhov", "obj", "default")
		require.NoError(t, cl.Create(ctx, &obj))

		// Act
		require.NoError(t, phase.Update(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("update on non-empty cloud creates resource", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, phase := setupAllocator(t)
		obj1 := createObject("sukhov", "obj1", "default")
		_, err := ad.Create(ctx, obj1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj1.ClusterName, obj1.Name))
		require.NoError(t, err)
		require.NoError(t, cl.Create(ctx, &obj1))
		obj2 := createObject("abdullah", "obj2", "other-namespace")
		_, err = ad.Create(ctx, obj2.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj2.ClusterName, obj2.Name))
		require.NoError(t, err)
		require.NoError(t, cl.Create(ctx, &obj2))

		obj3 := createObject("gulchatay", "obj3", "default")
		require.NoError(t, cl.Create(ctx, &obj3))

		// Act
		require.NoError(t, phase.Update(ctx, log, &obj3))
		upd, err := phase.IsUpdated(ctx, log, &obj3)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})
}

func TestAllocatorCleanup(t *testing.T) {
	t.Run("cleanup on cloud with resource deletes resource", func(t *testing.T) {
		// Arrange
		ctx, log, cl, _, phase := setupAllocator(t)
		obj1 := createObject("sukhov", "obj1", "default")
		require.NoError(t, cl.Create(ctx, &obj1))
		require.NoError(t, phase.Update(ctx, log, &obj1))

		obj2 := createObject("abdullah", "obj2", "other-namespace")
		require.NoError(t, cl.Create(ctx, &obj2))
		require.NoError(t, phase.Update(ctx, log, &obj2))

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj1))
		upd, err := phase.IsUpdated(ctx, log, &obj1)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("cleanup on cloud without resource does nothing", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, phase := setupAllocator(t)
		obj1 := createObject("sukhov", "obj1", "default")
		_, err := ad.Create(ctx, obj1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj1.ClusterName, obj1.Name))
		require.NoError(t, err)
		require.NoError(t, cl.Create(ctx, &obj1))
		require.NoError(t, phase.Update(ctx, log, &obj1))

		obj2 := createObject("abdullah", "obj2", "other-namespace")

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj2))
		upd, err := phase.IsUpdated(ctx, log, &obj2)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})
}
