// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/adapter"
	logrfake "k8s-connectors/testing/logr-fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestSpecMatcherIsUpdated(t *testing.T) {
	t.Run("is updated on matching spec", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		phase := SpecMatcher{
			Sdk: &ad,
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

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("is not updated on not matching spec", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		phase := SpecMatcher{
			Sdk: &ad,
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

		// Act
		obj.Spec.Name = "resource-upd"
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("attempt to change immutable field fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		phase := SpecMatcher{
			Sdk: &ad,
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
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		phase := SpecMatcher{
			Sdk: &ad,
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
