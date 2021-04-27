// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/adapter"
	"k8s-connectors/pkg/config"
	logrfake "k8s-connectors/testing/logr-fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestAllocatorIsUpdated(t *testing.T) {
	t.Run("is updated on empty cloud", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		phase := Allocator{
			Sdk: &ad,
		}
		obj := connectorsv1.YandexContainerRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj",
				Namespace: "default",
			},
		}
		// Act

		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("is updated on cloud with only this registry", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		phase := Allocator{
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
		res, err := ad.Create(ctx, &containerregistry.CreateRegistryRequest{
			FolderId: obj.Spec.FolderId,
			Name:     obj.Spec.Name,
			Labels: map[string]string{
				config.CloudClusterLabel: "",
				config.CloudNameLabel:    "obj",
			},
		})
		require.NoError(t, err)
		obj.Status.Id = res.Id
		// Act

		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("is updated on cloud with other registries", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		phase := Allocator{
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
		_, err := ad.Create(ctx, &containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "resource1",
			Labels: map[string]string{
				config.CloudClusterLabel: "",
				config.CloudNameLabel:    "obj1",
			},
		})
		require.NoError(t, err)
		_, err = ad.Create(ctx, &containerregistry.CreateRegistryRequest{
			FolderId: "other-folder",
			Name:     "resource2",
			Labels: map[string]string{
				config.CloudClusterLabel: "",
				config.CloudNameLabel:    "obj2",
			},
		})
		require.NoError(t, err)

		// Act

		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("is updated on cloud with this and other registries", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		ad := adapter.NewFakeYandexContainerRegistryAdapter()
		phase := Allocator{
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
		res, err := ad.Create(ctx, &containerregistry.CreateRegistryRequest{
			FolderId: obj.Spec.FolderId,
			Name:     obj.Spec.Name,
			Labels: map[string]string{
				config.CloudClusterLabel: "",
				config.CloudNameLabel:    "obj",
			},
		})
		require.NoError(t, err)
		obj.Status.Id = res.Id
		_, err = ad.Create(ctx, &containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "resource1",
			Labels: map[string]string{
				config.CloudClusterLabel: "",
				config.CloudNameLabel:    "obj1",
			},
		})
		require.NoError(t, err)
		_, err = ad.Create(ctx, &containerregistry.CreateRegistryRequest{
			FolderId: "other-folder",
			Name:     "resource2",
			Labels: map[string]string{
				config.CloudClusterLabel: "",
				config.CloudNameLabel:    "obj2",
			},
		})
		require.NoError(t, err)
		// Act

		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})
}

func TestAllocatorUpdate(t *testing.T) {
	t.Run("update on empty cloud creates resource", func(t *testing.T) {

	})

	t.Run("update on non-empty cloud creates resource", func(t *testing.T) {

	})
}

func TestAllocatorCleanup(t *testing.T) {
	t.Run("cleanup on cloud with resource deletes resource", func(t *testing.T) {

	})

	t.Run("cleanup on cloud with this and other resources deletes resources", func(t *testing.T) {

	})

	t.Run("cleanup on cloud without resource does nothing", func(t *testing.T) {

	})
}
