// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"

	connectorsv1 "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/api/v1"
	sakeyconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/pkg/config"
)

func TestUpdateStatus(t *testing.T) {
	t.Run(
		"update retains correct status", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("sukhov", "obj", "default")
			resp, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(rc.clusterID, obj.Name),
			)
			require.NoError(t, err)
			obj.Status.KeyID = resp.AccessKey.Id
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, rc.updateStatus(ctx, log, &obj, resp.AccessKey))
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
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("sukhov", "obj", "default")
			resp, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(rc.clusterID, obj.Name),
			)
			require.NoError(t, err)
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, rc.updateStatus(ctx, log, &obj, resp.AccessKey))
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
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("sukhov", "obj", "default")
			resp, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(rc.clusterID, obj.Name),
			)
			require.NoError(t, err)
			obj.Status.KeyID = "definitely-not-id"
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, rc.updateStatus(ctx, log, &obj, resp.AccessKey))
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
