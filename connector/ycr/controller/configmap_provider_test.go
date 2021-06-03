// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	ycrconfig "k8s-connectors/connector/ycr/pkg/config"
)

func createConfigMap(ctx context.Context, cl client.Client, t *testing.T, objectMetaName, namespace string) {
	require.NoError(
		t, cl.Create(
			ctx, &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ycrconfig.ShortName + "-" + objectMetaName + "-configmap",
					Namespace: namespace,
				},
			},
		),
	)
}

func TestProvideConfigmap(t *testing.T) {
	t.Run(
		"update on empty cloud creates configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl, _, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))
			require.NoError(t, rc.allocateResource(ctx, log, &obj))
			require.NoError(t, rc.updateStatus(ctx, log, &obj))

			// Act
			require.NoError(t, rc.provideConfigMap(ctx, log, &obj))
			var objUpd connectorsv1.YandexContainerRegistry
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "obj",
				Namespace: "default",
			}, &objUpd))
			var cmap v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "ycr-obj-configmap",
				Namespace: "default",
			}, &cmap))

			// Assert
			assert.Equal(t, objUpd.Status.ID, cmap.Data["ID"])
		},
	)

	t.Run(
		"update on non-empty cluster creates configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl, _, rc := setup(t)

			obj := createObject("resource", "folder", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))
			require.NoError(t, rc.allocateResource(ctx, log, &obj))
			require.NoError(t, rc.updateStatus(ctx, log, &obj))

			createConfigMap(ctx, cl, t, "obj1", "default")
			createConfigMap(ctx, cl, t, "obj", "other-namespace")

			// Act
			require.NoError(t, rc.provideConfigMap(ctx, log, &obj))
			var objUpd connectorsv1.YandexContainerRegistry
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "obj",
				Namespace: "default",
			}, &objUpd))
			var cmap v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "ycr-obj-configmap",
				Namespace: "default",
			}, &cmap))

			// Assert
			assert.Equal(t, objUpd.Status.ID, cmap.Data["ID"])
		},
	)
}

func TestRemoveConfigmap(t *testing.T) {
	t.Run(
		"cleanup on cloud with other configmaps does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, cl, _, rc := setup(t)

			obj := createObject("resource", "folder", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))
			require.NoError(t, rc.allocateResource(ctx, log, &obj))
			require.NoError(t, rc.updateStatus(ctx, log, &obj))
			require.NoError(t, rc.provideConfigMap(ctx, log, &obj))

			createConfigMap(ctx, cl, t, "obj1", "default")
			createConfigMap(ctx, cl, t, "obj", "other-namespace")

			// Act
			require.NoError(t, rc.removeConfigMap(ctx, log, &obj))
			var cmap v1.ConfigMap
			err := cl.Get(ctx, client.ObjectKey{
				Name:      "ycr-obj-configmap",
				Namespace: "default",
			}, &cmap)

			// Assert
			assert.Error(t, err)
			assert.True(t, errors.IsNotFound(err))
		},
	)

	t.Run(
		"cleanup on cloud with this and other configmaps deletes this configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl, _, rc := setup(t)

			obj := createObject("resource", "folder", "obj", "default")

			require.NoError(t, rc.provideConfigMap(ctx, log, &obj))

			createConfigMap(ctx, cl, t, "obj1", "default")
			createConfigMap(ctx, cl, t, "obj", "other-namespace")

			// Act
			require.NoError(t, rc.removeConfigMap(ctx, log, &obj))
			require.NoError(t, rc.removeConfigMap(ctx, log, &obj))
			var cmap v1.ConfigMap
			err := cl.Get(ctx, client.ObjectKey{
				Name:      "ycr-obj-configmap",
				Namespace: "default",
			}, &cmap)

			// Assert
			assert.Error(t, err)
			assert.True(t, errors.IsNotFound(err))
		},
	)

	t.Run(
		"cleanup on cloud with this configmap deletes this configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl, _, rc := setup(t)

			obj := createObject("resource", "folder", "obj", "default")

			require.NoError(t, rc.provideConfigMap(ctx, log, &obj))

			// Act
			require.NoError(t, rc.removeConfigMap(ctx, log, &obj))
			require.NoError(t, rc.removeConfigMap(ctx, log, &obj))
			var cmap v1.ConfigMap
			err := cl.Get(ctx, client.ObjectKey{
				Name:      "ycr-obj-configmap",
				Namespace: "default",
			}, &cmap)

			// Assert
			assert.Error(t, err)
			assert.True(t, errors.IsNotFound(err))
		},
	)
}
