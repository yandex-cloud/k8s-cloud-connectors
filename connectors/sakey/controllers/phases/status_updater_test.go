// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
	"k8s-connectors/connectors/sakey/controllers/adapter"
	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupStatusUpdater(t *testing.T) (
	context.Context, logr.Logger, client.Client, adapter.StaticAccessKeyAdapter, StaticAccessKeyPhase,
) {
	ad := adapter.NewFakeStaticAccessKeyAdapter()
	cl := k8sfake.NewFakeClient()
	return context.Background(), logrfake.NewFakeLogger(t), cl, &ad, &StatusUpdater{
		Sdk:    &ad,
		Client: &cl,
	}
}

func TestStatusUpdaterUpdate(t *testing.T) {
	t.Run(
		"update retains correct status", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, phase := setupStatusUpdater(t)
			obj := createObject("sukhov", "obj", "default")
			resp, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj.ClusterName, obj.Name),
			)
			require.NoError(t, err)
			obj.Status.KeyID = resp.AccessKey.Id
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, phase.Update(ctx, log, &obj))
			var current connectorsv1.StaticAccessKey
			require.NoError(
				t, cl.Get(
					ctx, types.NamespacedName{
						Namespace: "default",
						Name:      "obj",
					}, &current,
				),
			)

			// Assert
			assert.Equal(t, obj.Status.KeyID, current.Status.KeyID)
		},
	)

	t.Run(
		"update fills empty status", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, phase := setupStatusUpdater(t)
			obj := createObject("sukhov", "obj", "default")
			resp, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj.ClusterName, obj.Name),
			)
			require.NoError(t, err)
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, phase.Update(ctx, log, &obj))
			var current connectorsv1.StaticAccessKey
			require.NoError(
				t, cl.Get(
					ctx, types.NamespacedName{
						Namespace: "default",
						Name:      "obj",
					}, &current,
				),
			)

			// Assert
			assert.Equal(t, resp.AccessKey.Id, current.Status.KeyID)
		},
	)

	t.Run(
		"update updates incorrect status", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, phase := setupStatusUpdater(t)
			obj := createObject("sukhov", "obj", "default")
			resp, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj.ClusterName, obj.Name),
			)
			require.NoError(t, err)
			obj.Status.KeyID = "definitely-not-id"
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, phase.Update(ctx, log, &obj))
			var current connectorsv1.StaticAccessKey
			require.NoError(
				t, cl.Get(
					ctx, types.NamespacedName{
						Namespace: "default",
						Name:      "obj",
					}, &current,
				),
			)

			// Assert
			assert.Equal(t, resp.AccessKey.Id, current.Status.KeyID)
		},
	)
}
