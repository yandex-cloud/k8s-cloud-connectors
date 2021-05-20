// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	"k8s-connectors/connector/ycr/controller/adapter"
	ycrutils "k8s-connectors/connector/ycr/pkg/util"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupStatusUpdater(t *testing.T) (
	context.Context, logr.Logger, client.Client, adapter.YandexContainerRegistryAdapter, YandexContainerRegistryPhase,
) {
	cl := k8sfake.NewFakeClient()
	ad := adapter.NewFakeYandexContainerRegistryAdapter()
	return context.Background(), logrfake.NewFakeLogger(t), cl, &ad, &StatusUpdater{
		Sdk:    &ad,
		Client: cl,
	}
}

func TestStatusUpdaterUpdate(t *testing.T) {
	t.Run(
		"update retains matching status", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, phase := setupStatusUpdater(t)
			obj := createObject("resource", "folder", "obj", "default")
			createResourceRequireNoError(ctx, ad, t, obj.Spec.Name, obj.Spec.FolderID, obj.Name, obj.ClusterName)

			res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "", ad)
			require.NoError(t, err)
			obj.Status.ID = res1.Id
			obj.Status.Labels = res1.Labels
			obj.Status.CreatedAt = res1.CreatedAt.String()

			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, phase.Update(ctx, log, &obj))
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
			ctx, log, cl, ad, phase := setupStatusUpdater(t)
			obj := createObject("resource", "folder", "obj", "default")
			createResourceRequireNoError(ctx, ad, t, obj.Spec.Name, obj.Spec.FolderID, obj.Name, obj.ClusterName)

			res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "", ad)
			require.NoError(t, err)
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, phase.Update(ctx, log, &obj))
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
			ctx, log, cl, ad, phase := setupStatusUpdater(t)
			obj := createObject("resource", "folder", "obj", "default")
			createResourceRequireNoError(ctx, ad, t, obj.Spec.Name, obj.Spec.FolderID, obj.Name, obj.ClusterName)

			res1, err := ycrutils.GetRegistry(ctx, "", "folder", "obj", "", ad)
			require.NoError(t, err)
			obj.Status.ID = "definitely-not-id"
			obj.Status.Labels = map[string]string{"key": "label"}
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, phase.Update(ctx, log, &obj))
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
