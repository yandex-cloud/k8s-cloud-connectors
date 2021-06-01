// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	ycrutils "k8s-connectors/connector/ycr/pkg/util"
)

func TestStatusUpdaterUpdate(t *testing.T) {
	t.Run(
		"update retains matching status", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")
			require.NoError(t, rc.allocateResource(ctx, log, &obj))

			res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "test-cluster", ad)
			require.NoError(t, err)
			obj.Status.ID = res1.Id
			obj.Status.Labels = res1.Labels
			obj.Status.CreatedAt = res1.CreatedAt.String()

			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, rc.updateStatus(ctx, log, &obj))
			var current connectorsv1.YandexContainerRegistry
			require.NoError(
				t, cl.Get(
					ctx, types.NamespacedName{
						Namespace: "default",
						Name:      "obj",
					}, &current,
				),
			)

			// Assert
			assert.Equal(t, res1.Id, current.Status.ID)
			assert.Equal(t, res1.Labels, current.Status.Labels)
			assert.Equal(t, res1.CreatedAt.String(), current.Status.CreatedAt)
		},
	)

	t.Run(
		"update matches empty status", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")
			require.NoError(t, rc.allocateResource(ctx, log, &obj))

			res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "test-cluster", ad)
			require.NoError(t, err)
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, rc.updateStatus(ctx, log, &obj))
			var current connectorsv1.YandexContainerRegistry
			require.NoError(
				t, cl.Get(
					ctx, types.NamespacedName{
						Namespace: "default",
						Name:      "obj",
					}, &current,
				),
			)

			// Assert
			assert.Equal(t, res1.Id, current.Status.ID)
			assert.Equal(t, res1.Labels, current.Status.Labels)
			assert.Equal(t, res1.CreatedAt.String(), current.Status.CreatedAt)
		},
	)

	t.Run(
		"update matches non-matching status", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")
			require.NoError(t, rc.allocateResource(ctx, log, &obj))

			res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "test-cluster", ad)
			require.NoError(t, err)
			obj.Status.ID = "definitely-not-id"
			obj.Status.Labels = map[string]string{"key": "label"}
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, rc.updateStatus(ctx, log, &obj))
			var current connectorsv1.YandexContainerRegistry
			require.NoError(
				t, cl.Get(
					ctx, types.NamespacedName{
						Namespace: "default",
						Name:      "obj",
					}, &current,
				),
			)

			// Assert
			assert.Equal(t, res1.Id, current.Status.ID)
			assert.Equal(t, res1.Labels, current.Status.Labels)
			assert.Equal(t, res1.CreatedAt.String(), current.Status.CreatedAt)
		},
	)
}
