// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrconfig "k8s-connectors/connectors/ycr/pkg/config"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestEndpointProviderIsUpdated(t *testing.T) {
	t.Run("is updated on configmap existence", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj-configmap",
				Namespace: "default",
			},
		}))

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("is updated on many configmap existence", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj-configmap",
				Namespace: "default",
			},
		}))
		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj1-configmap",
				Namespace: "default",
			},
		}))
		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj-configmap",
				Namespace: "other-namespace",
			},
		}))

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("is not updated on empty cloud", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})

	t.Run("is not updated on other objects existence", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj1-configmap",
				Namespace: "default",
			},
		}))
		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj-configmap",
				Namespace: "other-namespace",
			},
		}))

		// Act
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.False(t, upd)
	})
}

func TestEndpointProviderUpdate(t *testing.T) {
	t.Run("update on empty cloud creates configmap", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		// Act
		require.NoError(t, phase.Update(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})

	t.Run("update on non-empty cloud creates configmap", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj1-configmap",
				Namespace: "default",
			},
		}))
		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj-configmap",
				Namespace: "other-namespace",
			},
		}))

		// Act
		require.NoError(t, phase.Update(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// Assert
		assert.True(t, upd)
	})
}

func TestEndpointProviderCleanup(t *testing.T) {
	t.Run("cleanup on empty cloud does nothing", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// TODO (covariance) implement fake List method
		// var lst v1.ConfigMapList
		// require.NoError(t, cl.List(ctx, &lst))

		// Assert
		assert.False(t, upd)
		// assert.Len(t, lst.Items, 0)
	})

	t.Run("cleanup on cloud with other configmaps does nothing", func(t *testing.T) {
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj1-configmap",
				Namespace: "default",
			},
		}))
		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj-configmap",
				Namespace: "other-namespace",
			},
		}))

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// TODO (covariance) implement fake List method
		// var lst v1.ConfigMapList
		// require.NoError(t, cl.List(ctx, &lst))

		// Assert
		assert.False(t, upd)
		// assert.Len(t, lst.Items, 2)
	})

	t.Run("cleanup on cloud with this and other configmaps deletes this configmap", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		require.NoError(t, phase.Update(ctx, log, &obj))

		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj1-configmap",
				Namespace: "default",
			},
		}))
		require.NoError(t, cl.Create(ctx, &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ycrconfig.ShortName + "-obj-configmap",
				Namespace: "other-namespace",
			},
		}))

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// TODO (covariance) implement fake List method
		// var lst v1.ConfigMapList
		// require.NoError(t, cl.List(ctx, &lst))

		// Assert
		assert.False(t, upd)
		// assert.Len(t, lst.Items, 2)
	})

	t.Run("cleanup on cloud with this configmap deletes this configmap", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		log := logrfake.NewFakeLogger(t)
		cl := k8sfake.NewFakeClient()
		phase := EndpointProvider{
			Client: &cl,
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

		require.NoError(t, phase.Update(ctx, log, &obj))

		// Act
		require.NoError(t, phase.Cleanup(ctx, log, &obj))
		upd, err := phase.IsUpdated(ctx, log, &obj)
		require.NoError(t, err)

		// TODO (covariance) implement fake List method
		// var lst v1.ConfigMapList
		// require.NoError(t, cl.List(ctx, &lst))

		// Assert
		assert.False(t, upd)
		// assert.Len(t, lst.Items, 0)
	})
}
