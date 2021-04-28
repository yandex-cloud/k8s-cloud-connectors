// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrconfig "k8s-connectors/connectors/ycr/pkg/config"
	"k8s-connectors/pkg/utils"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func setupFinalizerRegistrar(t *testing.T) (context.Context, logr.Logger, client.Client, YandexContainerRegistryPhase) {
	cl := k8sfake.NewFakeClient()
	return context.Background(), logrfake.NewFakeLogger(t), cl, &FinalizerRegistrar{Client: &cl}
}

func createObjectWithFinalizers(ctx context.Context, cl client.Client, t *testing.T, metaName, namespace string, finalizers []string) *connectorsv1.YandexContainerRegistry {
	obj := connectorsv1.YandexContainerRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:       metaName,
			Namespace:  namespace,
			Finalizers: finalizers,
		},
	}
	require.NoError(t, cl.Create(ctx, &obj))
	return &obj
}

func TestFinalizerRegistrarIsUpdated(t *testing.T) {
	t.Run("empty finalizers means not updated", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{})

		// Act
		updated, err := phase.IsUpdated(context.Background(), log, obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, updated)
	})

	t.Run("other finalizers means not updated", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{"not.that.finalizer", "yet.another.false.finalizer"})

		// Act
		updated, err := phase.IsUpdated(context.Background(), log, obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, updated)
	})

	t.Run("finalizer exist means updated", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{ycrconfig.FinalizerName})

		// Act
		updated, err := phase.IsUpdated(context.Background(), log, obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, updated)
	})

	t.Run("finalizer and others exist means updated", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{"not.that.finalizer", ycrconfig.FinalizerName, "yet.another.false.finalizer"})

		// Act
		updated, err := phase.IsUpdated(context.Background(), log, obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, updated)
	})
}

func TestFinalizerRegistrarUpdate(t *testing.T) {
	t.Run("update on empty finalizer list adds finalizer", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{})

		// Act
		require.NoError(t, phase.Update(context.Background(), log, obj))
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, cl.Get(context.Background(), utils.NamespacedName(obj), &res))

		// Assert
		assert.Len(t, res.Finalizers, 1)
		assert.Contains(t, res.Finalizers, ycrconfig.FinalizerName)
	})

	t.Run("update on non-empty finalizer list adds finalizer", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{"not.that.finalizer", "yet.another.finalizer"})

		// Act
		require.NoError(t, phase.Update(context.Background(), log, obj))
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, cl.Get(context.Background(), utils.NamespacedName(obj), &res))

		// Assert
		assert.Len(t, res.Finalizers, 3)
		assert.Contains(t, res.Finalizers, "not.that.finalizer")
		assert.Contains(t, res.Finalizers, "yet.another.finalizer")
		assert.Contains(t, res.Finalizers, ycrconfig.FinalizerName)
	})
}

func TestFinalizerRegistrarCleanup(t *testing.T) {
	t.Run("cleanup on empty finalizer list does nothing", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{})

		// Act
		require.NoError(t, phase.Cleanup(context.Background(), log, obj))
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, cl.Get(context.Background(), utils.NamespacedName(obj), &res))

		// Assert
		assert.Len(t, res.Finalizers, 0)
	})

	t.Run("cleanup on non-empty finalizer list removes finalizer", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{"not.that.finalizer", ycrconfig.FinalizerName, "yet.another.finalizer"})

		// Act
		require.NoError(t, phase.Cleanup(context.Background(), log, obj))
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, cl.Get(context.Background(), utils.NamespacedName(obj), &res))

		// Assert
		assert.Len(t, res.Finalizers, 2)
		assert.Contains(t, res.Finalizers, "not.that.finalizer")
		assert.Contains(t, res.Finalizers, "yet.another.finalizer")
	})

	t.Run("cleanup on non-empty finalizer list without needed finalizer does nothing", func(t *testing.T) {
		// Arrange
		ctx, log, cl, phase := setupFinalizerRegistrar(t)
		obj := createObjectWithFinalizers(ctx, cl, t,"obj", "default", []string{"not.that.finalizer", "yet.another.finalizer"})

		// Act
		require.NoError(t, phase.Cleanup(context.Background(), log, obj))
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, cl.Get(context.Background(), utils.NamespacedName(obj), &res))

		// Assert
		assert.Len(t, res.Finalizers, 2)
		assert.Contains(t, res.Finalizers, "not.that.finalizer")
		assert.Contains(t, res.Finalizers, "yet.another.finalizer")
	})
}
