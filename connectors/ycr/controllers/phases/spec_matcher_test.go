// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s-connectors/connectors/ycr/controllers/adapter"
	logrfake "k8s-connectors/testing/logr-fake"
	"testing"
)

func setupSpecMatcher(t *testing.T) (context.Context, logr.Logger, adapter.YandexContainerRegistryAdapter, YandexContainerRegistryPhase, Allocator) {
	ad := adapter.NewFakeYandexContainerRegistryAdapter()
	return context.Background(), logrfake.NewFakeLogger(t), &ad, &SpecMatcher{Sdk: &ad}, Allocator{Sdk: &ad}
}

func TestSpecMatcherIsUpdated(t *testing.T) {
	t.Run("is updated on matching spec", func(t *testing.T) {
		// Arrange
		ctx, log, _, phase, allocator := setupSpecMatcher(t)
		obj := CreateObject("resource", "folder", "obj", "default")
		require.NoError(t, allocator.Update(ctx, log, &obj))

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("is not updated on not matching spec", func(t *testing.T) {
		// Arrange
		ctx, log, _, phase, allocator := setupSpecMatcher(t)
		obj := CreateObject("resource", "folder", "obj", "default")
		require.NoError(t, allocator.Update(ctx, log, &obj))

		// Act
		obj.Spec.Name = "resource-upd"
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("attempt to change immutable field fails", func(t *testing.T) {
		// Arrange
		ctx, log, _, phase, allocator := setupSpecMatcher(t)
		obj := CreateObject("resource", "folder", "obj", "default")
		require.NoError(t, allocator.Update(ctx, log, &obj))

		// Act
		obj.Spec.FolderId = "other-folder"
		_, err := phase.IsUpdated(ctx, log, &obj)

		// Assert
		assert.Error(t, err)
	})
}

func TestSpecMatcherUpdate(t *testing.T) {
	t.Run("update matches cloud object with spec of resource", func(t *testing.T) {
		// Arrange
		ctx, log, _, phase, allocator := setupSpecMatcher(t)
		obj := CreateObject("resource", "folder", "obj", "default")
		require.NoError(t, allocator.Update(ctx, log, &obj))

		// Act
		obj.Spec.Name = "resource-upd"
		require.NoError(t, phase.Update(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})
}

func TestSpecMatcherCleanup(t *testing.T) {
	// There's nothing to do in cleanup for this phase
}
