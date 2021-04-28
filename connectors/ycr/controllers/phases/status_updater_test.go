// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/adapter"
	ycrutils "k8s-connectors/connectors/ycr/pkg/util"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestStatusUpdaterIsUpdated(t *testing.T) {
	// This phase must never be updated
}

func TestStatusUpdaterUpdate(t *testing.T) {
	t.Run("update retains matching status", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		cl := k8sfake.NewFakeClient()
		phase := StatusUpdater{
			Sdk:    &ad,
			Client: &cl,
		}
		allocator := Allocator{
			Sdk: &ad,
		}
		obj := connectorsv1.YandexContainerRegistry{
			Spec: connectorsv1.YandexContainerRegistrySpec{
				Name:     "resource",
				FolderId: "folder",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj",
				Namespace: "default",
			},
		}
		require.NoError(t, allocator.Update(ctx, log, &obj))

		res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "", &ad)
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
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		cl := k8sfake.NewFakeClient()
		phase := StatusUpdater{
			Sdk:    &ad,
			Client: &cl,
		}
		allocator := Allocator{
			Sdk: &ad,
		}
		obj := connectorsv1.YandexContainerRegistry{
			Spec: connectorsv1.YandexContainerRegistrySpec{
				Name:     "resource",
				FolderId: "folder",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj",
				Namespace: "default",
			},
		}
		require.NoError(t, allocator.Update(ctx, log, &obj))

		res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "", &ad)
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
