// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sakeyconfig "github.com/yandex-cloud/k8s-cloud-connectors/connector/sakey/pkg/config"
)

func TestAllocate(t *testing.T) {
	t.Run(
		"allocate on empty cloud creates resource", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("sukhov", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))

			// Act
			res, err := rc.allocateResource(ctx, log, &obj)
			require.NoError(t, err)
			lst, err := ad.List(ctx, "sukhov")
			require.NoError(t, err)
			var secret v1.Secret
			require.NoError(t, cl.Get(ctx, client.ObjectKey{
				Namespace: "default",
				Name:      obj.Status.SecretName,
			}, &secret))

			// Assert
			assert.Len(t, lst, 1)

			// Check match with cloud
			assert.Equal(t, "sukhov", lst[0].ServiceAccountId)
			assert.Equal(t, sakeyconfig.GetStaticAccessKeyDescription("test-cluster", "obj"), lst[0].Description)
			// Check values in the secret
			assert.Equal(t, lst[0].Id, string(secret.Data["secret"]))
			assert.Equal(t, lst[0].Id, string(secret.Data["key"]))
			// Check match with returned object
			assert.Equal(t, "sukhov", res.ServiceAccountId)
			assert.Equal(t, sakeyconfig.GetStaticAccessKeyDescription("test-cluster", "obj"), res.Description)
			assert.Equal(t, string(secret.Data["key"]), res.KeyId)
		},
	)

	t.Run(
		"allocate on non-empty cloud creates resource", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
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
			res, err := rc.allocateResource(ctx, log, &obj3)
			require.NoError(t, err)
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

			// Check match with cloud
			assert.Equal(t, "gulchatay", lst3[0].ServiceAccountId)
			assert.Equal(t, sakeyconfig.GetStaticAccessKeyDescription("test-cluster", "obj3"), lst3[0].Description)
			// Check values in the secret
			assert.Equal(t, lst3[0].Id, string(secret.Data["secret"]))
			assert.Equal(t, lst3[0].Id, string(secret.Data["key"]))
			// Check match with returned object
			assert.Equal(t, "gulchatay", res.ServiceAccountId)
			assert.Equal(t, sakeyconfig.GetStaticAccessKeyDescription("test-cluster", "obj3"), res.Description)
			assert.Equal(t, string(secret.Data["key"]), res.KeyId)
		},
	)
}

func TestDeallocate(t *testing.T) {
	t.Run(
		"deallocate on cloud with resource deletes resource", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj1 := createObject("sukhov", "obj1", "default")
			require.NoError(t, cl.Create(ctx, &obj1))
			_, err := rc.allocateResource(ctx, log, &obj1)
			require.NoError(t, err)

			obj2 := createObject("abdullah", "obj2", "other-namespace")
			require.NoError(t, cl.Create(ctx, &obj2))
			_, err = rc.allocateResource(ctx, log, &obj2)
			require.NoError(t, err)

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
		"deallocate on cloud without resource does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj1 := createObject("sukhov", "obj1", "default")
			require.NoError(t, cl.Create(ctx, &obj1))
			_, err := rc.allocateResource(ctx, log, &obj1)
			require.NoError(t, err)

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
