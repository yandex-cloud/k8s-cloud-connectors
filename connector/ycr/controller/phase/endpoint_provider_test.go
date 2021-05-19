// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ycrconfig "k8s-connectors/connector/ycr/pkg/config"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupEndpointProvider(t *testing.T) (context.Context, logr.Logger, client.Client, YandexContainerRegistryPhase) {
	cl := k8sfake.NewFakeClient()
	return context.Background(), logrfake.NewFakeLogger(t), cl, &EndpointProvider{Client: cl}
}

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

func TestEndpointProviderIsUpdated(t *testing.T) {
	t.Run(
		"is updated on configmap existence", func(t *testing.T) {
			// Arrange
			ctx, log, cl, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")
			createConfigMap(ctx, cl, t, "obj", "default")

			// Act
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.True(t, upd)
		},
	)

	t.Run(
		"is updated on many configmap existence", func(t *testing.T) {
			// Arrange
			ctx, log, cl, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")
			createConfigMap(ctx, cl, t, "obj", "default")

			createConfigMap(ctx, cl, t, "obj1", "default")
			createConfigMap(ctx, cl, t, "obj", "other-namespace")

			// Act
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.True(t, upd)
		},
	)

	t.Run(
		"is not updated on empty cloud", func(t *testing.T) {
			// Arrange
			ctx, log, _, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")

			// Act
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.False(t, upd)
		},
	)

	t.Run(
		"is not updated on other objects existence", func(t *testing.T) {
			// Arrange
			ctx, log, cl, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")

			createConfigMap(ctx, cl, t, "obj1", "default")
			createConfigMap(ctx, cl, t, "obj", "other-namespace")

			// Act
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.False(t, upd)
		},
	)
}

func TestEndpointProviderUpdate(t *testing.T) {
	t.Run(
		"update on empty cloud creates configmap", func(t *testing.T) {
			// Arrange
			ctx, log, _, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")

			// Act
			require.NoError(t, phase.Update(ctx, log, &obj))
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.True(t, upd)
		},
	)

	t.Run(
		"update on non-empty cloud creates configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")

			createConfigMap(ctx, cl, t, "obj1", "default")
			createConfigMap(ctx, cl, t, "obj", "other-namespace")

			// Act
			require.NoError(t, phase.Update(ctx, log, &obj))
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.True(t, upd)
		},
	)
}

func TestEndpointProviderCleanup(t *testing.T) {
	t.Run(
		"cleanup on empty cloud does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, _, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")

			// Act
			require.NoError(t, phase.Cleanup(ctx, log, &obj))
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.False(t, upd)
		},
	)

	t.Run(
		"cleanup on cloud with other configmaps does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, cl, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")

			createConfigMap(ctx, cl, t, "obj1", "default")
			createConfigMap(ctx, cl, t, "obj", "other-namespace")

			// Act
			require.NoError(t, phase.Cleanup(ctx, log, &obj))
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.False(t, upd)
		},
	)

	t.Run(
		"cleanup on cloud with this and other configmaps deletes this configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")

			require.NoError(t, phase.Update(ctx, log, &obj))

			createConfigMap(ctx, cl, t, "obj1", "default")
			createConfigMap(ctx, cl, t, "obj", "other-namespace")

			// Act
			require.NoError(t, phase.Cleanup(ctx, log, &obj))
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.False(t, upd)
		},
	)

	t.Run(
		"cleanup on cloud with this configmap deletes this configmap", func(t *testing.T) {
			// Arrange
			ctx, log, _, phase := setupEndpointProvider(t)

			obj := createObject("resource", "folder", "obj", "default")

			require.NoError(t, phase.Update(ctx, log, &obj))

			// Act
			require.NoError(t, phase.Cleanup(ctx, log, &obj))
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.False(t, upd)
		},
	)
}
