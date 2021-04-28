// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s-connectors/connectors/ycr/controllers/adapter"
	ycrutils "k8s-connectors/connectors/ycr/pkg/util"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestStatusUpdaterIsUpdated(t *testing.T) {
	// This phase must never be updated
}

func setupStatusUpdater(t *testing.T) (context.Context, logr.Logger, client.Client, adapter.YandexContainerRegistryAdapter, YandexContainerRegistryPhase, Allocator) {
	cl := k8sfake.NewFakeClient()
	ad := adapter.NewFakeYandexContainerRegistryAdapter()
	return context.Background(), logrfake.NewFakeLogger(t), cl, &ad, &StatusUpdater{
		Sdk:    &ad,
		Client: &cl,
	}, Allocator{Sdk: &ad}
}

func TestStatusUpdaterUpdate(t *testing.T) {
	t.Run("update retains matching status", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, phase, allocator := setupStatusUpdater(t)
		obj := CreateObject("resource", "folder", "obj", "default")
		require.NoError(t, allocator.Update(ctx, log, &obj))

		res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "", ad)
		require.NoError(t, err)
		obj.Status.Id = res1.Id
		obj.Status.Labels = res1.Labels
		obj.Status.CreatedAt = res1.CreatedAt.String()

		require.NoError(t, cl.Create(ctx, &obj))

		// Act
		require.NoError(t, phase.Update(ctx, log, &obj))

		// Assert
		assert.Equal(t, res1.Id, obj.Status.Id)
		assert.Equal(t, res1.Labels, obj.Status.Labels)
		assert.Equal(t, res1.CreatedAt.String(), obj.Status.CreatedAt)
	})

	t.Run("update matches non-matching status", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, phase, allocator := setupStatusUpdater(t)
		obj := CreateObject("resource", "folder", "obj", "default")
		require.NoError(t, allocator.Update(ctx, log, &obj))

		res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "", ad)
		require.NoError(t, err)
		require.NoError(t, cl.Create(ctx, &obj))

		// Act
		require.NoError(t, phase.Update(ctx, log, &obj))

		// Assert
		assert.Equal(t, res1.Id, obj.Status.Id)
		assert.Equal(t, res1.Labels, obj.Status.Labels)
		assert.Equal(t, res1.CreatedAt.String(), obj.Status.CreatedAt)
	})
}

func TestStatusUpdaterCleanup(t *testing.T) {
	// There's nothing to do in cleanup for this phase
}
