// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s-connectors/connector/sakey/controller/adapter"
	sakeyconfig "k8s-connectors/connector/sakey/pkg/config"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupAllocator(t *testing.T) (
	context.Context, logr.Logger, client.Client, adapter.StaticAccessKeyAdapter, staticAccessKeyReconciler,
) {
	ad := adapter.NewFakeStaticAccessKeyAdapter()
	cl := k8sfake.NewFakeClient()
	log := logrfake.NewFakeLogger(t)
	return context.Background(), log, cl, &ad, staticAccessKeyReconciler{
		cl,
		&ad,
		log,
		"test-cluster",
	}
}

func TestAllocate(t *testing.T) {
	t.Run(
		"update on empty cloud creates resource", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setupAllocator(t)
			obj := createObject("sukhov", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			require.NoError(t, rc.allocateResource(ctx, log, &obj))
			lst, err := ad.List(ctx, "sukhov")
			require.NoError(t, err)
			var secret v1.Secret
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      obj.Status.SecretName,
			}, &secret))

			// Assert
			assert.Len(t, lst, 1)
			assert.Equal(t, "sukhov", lst[0].ServiceAccountId)
			assert.Equal(t, sakeyconfig.GetStaticAccessKeyDescription("test-cluster", "obj"), lst[0].Description)
			assert.Equal(t, lst[0].Id, string(secret.Data["secret"]))
			assert.Equal(t, lst[0].Id, string(secret.Data["key"]))
		},
	)

	t.Run(
		"update on non-empty cloud creates resource", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setupAllocator(t)
			obj1 := createObject("sukhov", "obj1", "default")
			_, err := ad.Create(
				ctx, obj1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj1.ClusterName, obj1.Name),
			)
			require.NoError(t, err)
			require.NoError(t, cl.Create(ctx, &obj1))
			obj2 := createObject("abdullah", "obj2", "other-namespace")
			_, err = ad.Create(
				ctx, obj2.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj2.ClusterName, obj2.Name),
			)
			require.NoError(t, err)
			require.NoError(t, cl.Create(ctx, &obj2))

			obj3 := createObject("gulchatay", "obj3", "default")
			require.NoError(t, cl.Create(ctx, &obj3))

			// Act
			require.NoError(t, rc.allocateResource(ctx, log, &obj3))
			lst1, err := ad.List(ctx, "sukhov")
			require.NoError(t, err)
			lst2, err := ad.List(ctx, "abdullah")
			require.NoError(t, err)
			lst3, err := ad.List(ctx, "gulchatay")
			require.NoError(t, err)
			var secret v1.Secret
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      obj3.Status.SecretName,
			}, &secret))

			// Assert
			assert.Len(t, lst1, 1)
			assert.Len(t, lst2, 1)
			assert.Len(t, lst3, 1)
			assert.Equal(t, "gulchatay", lst3[0].ServiceAccountId)
			assert.Equal(t, sakeyconfig.GetStaticAccessKeyDescription("test-cluster", "obj3"), lst3[0].Description)
			assert.Equal(t, lst3[0].Id, string(secret.Data["secret"]))
			assert.Equal(t, lst3[0].Id, string(secret.Data["key"]))
		},
	)
}

func TestDeallocate(t *testing.T) {
	t.Run(
		"cleanup on cloud with resource deletes resource", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setupAllocator(t)
			obj1 := createObject("sukhov", "obj1", "default")
			require.NoError(t, cl.Create(ctx, &obj1))
			require.NoError(t, rc.allocateResource(ctx, log, &obj1))

			obj2 := createObject("abdullah", "obj2", "other-namespace")
			require.NoError(t, cl.Create(ctx, &obj2))
			require.NoError(t, rc.allocateResource(ctx, log, &obj2))

			// Act
			require.NoError(t, rc.deallocateResource(ctx, log, &obj1))
			lst1, err := ad.List(ctx, "sukhov")
			require.NoError(t, err)
			lst2, err := ad.List(ctx, "abdullah")
			require.NoError(t, err)

			// Assert
			assert.Len(t, lst1, 0)
			assert.Len(t, lst2, 1)
		},
	)

	t.Run(
		"cleanup on cloud without resource does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setupAllocator(t)
			obj1 := createObject("sukhov", "obj1", "default")
			require.NoError(t, cl.Create(ctx, &obj1))
			require.NoError(t, rc.allocateResource(ctx, log, &obj1))

			obj2 := createObject("abdullah", "obj2", "other-namespace")

			// Act
			require.NoError(t, rc.deallocateResource(ctx, log, &obj2))
			lst1, err := ad.List(ctx, "sukhov")
			require.NoError(t, err)
			lst2, err := ad.List(ctx, "abdullah")
			require.NoError(t, err)

			// Assert
			assert.Len(t, lst1, 1)
			assert.Len(t, lst2, 0)
		},
	)
}
