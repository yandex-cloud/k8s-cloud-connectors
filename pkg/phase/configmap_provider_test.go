// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createConfigmap(
	ctx context.Context,
	cl client.Client,
	name, namespace string,
	contents map[string]string,
) error {
	return cl.Create(ctx, &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: contents,
	})
}

func TestProvideConfigmap(t *testing.T) {
	t.Run(
		"provide on empty cloud creates configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)

			// Act
			require.NoError(t, ProvideConfigmap(
				ctx,
				cl,
				log,
				"object",
				"kind",
				"default",
				map[string]string{"john": "dow"},
			))
			var res v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "kind-object-configmap",
				Namespace: "default",
			}, &res))

			// Assert
			assert.Equal(t, "dow", res.Data["john"])
		},
	)

	t.Run(
		"provide on non-empty cluster creates configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			require.NoError(t, createConfigmap(ctx, cl, "other-object", "default", map[string]string{"john": "snow"}))
			require.NoError(t, createConfigmap(ctx, cl, "object", "other-default", map[string]string{"john": "crow"}))

			// Act
			require.NoError(t, ProvideConfigmap(
				ctx,
				cl,
				log,
				"object",
				"kind",
				"default",
				map[string]string{"john": "dow"},
			))
			var res v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "kind-object-configmap",
				Namespace: "default",
			}, &res))

			// Check that we have not messed anything else up
			var kingInTheNorth v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "other-object",
				Namespace: "default",
			}, &kingInTheNorth))
			var nightWatcher v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "object",
				Namespace: "other-default",
			}, &nightWatcher))

			// Assert
			assert.Equal(t, "dow", res.Data["john"])
			assert.Equal(t, "snow", kingInTheNorth.Data["john"])
			assert.Equal(t, "crow", nightWatcher.Data["john"])
		},
	)
}

func TestRemoveConfigmap(t *testing.T) {
	t.Run(
		"remove on cluster with other configmaps does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			require.NoError(t, createConfigmap(ctx, cl, "other-object", "default", map[string]string{"john": "snow"}))
			require.NoError(t, createConfigmap(ctx, cl, "object", "other-default", map[string]string{"john": "crow"}))

			// Act
			require.NoError(t, RemoveConfigmap(
				ctx,
				cl,
				log,
				"object",
				"kind",
				"default",
			))

			var res v1.ConfigMap
			err := cl.Get(ctx, client.ObjectKey{
				Name:      "kind-object-configmap",
				Namespace: "default",
			}, &res)

			// Check that we have not messed anything else up
			var kingInTheNorth v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "other-object",
				Namespace: "default",
			}, &kingInTheNorth))
			var nightWatcher v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "object",
				Namespace: "other-default",
			}, &nightWatcher))

			// Assert
			assert.True(t, errors.IsNotFound(err))
			assert.Equal(t, "snow", kingInTheNorth.Data["john"])
			assert.Equal(t, "crow", nightWatcher.Data["john"])
		},
	)

	t.Run(
		"remove on cluster with this and other configmaps deletes this configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			require.NoError(t, createConfigmap(ctx, cl, "other-object", "default", map[string]string{"john": "snow"}))
			require.NoError(t, createConfigmap(ctx, cl, "object", "other-default", map[string]string{"john": "crow"}))
			require.NoError(t, ProvideConfigmap(
				ctx,
				cl,
				log,
				"object",
				"kind",
				"default",
				map[string]string{"john": "dow"},
			))

			// Act
			require.NoError(t, RemoveConfigmap(
				ctx,
				cl,
				log,
				"object",
				"kind",
				"default",
			))

			var res v1.ConfigMap
			err := cl.Get(ctx, client.ObjectKey{
				Name:      "kind-object-configmap",
				Namespace: "default",
			}, &res)

			// Check that we have not messed anything else up
			var kingInTheNorth v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "other-object",
				Namespace: "default",
			}, &kingInTheNorth))
			var nightWatcher v1.ConfigMap
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Name:      "object",
				Namespace: "other-default",
			}, &nightWatcher))

			// Assert
			assert.True(t, errors.IsNotFound(err))
			assert.Equal(t, "snow", kingInTheNorth.Data["john"])
			assert.Equal(t, "crow", nightWatcher.Data["john"])
		},
	)

	t.Run(
		"remove on cluster with this configmap deletes this configmap", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			require.NoError(t, ProvideConfigmap(
				ctx,
				cl,
				log,
				"object",
				"kind",
				"default",
				map[string]string{"john": "dow"},
			))

			// Act
			require.NoError(t, RemoveConfigmap(
				ctx,
				cl,
				log,
				"object",
				"kind",
				"default",
			))

			var res v1.ConfigMap
			err := cl.Get(ctx, client.ObjectKey{
				Name:      "kind-object-configmap",
				Namespace: "default",
			}, &res)
			// Assert
			assert.True(t, errors.IsNotFound(err))
		},
	)
}
