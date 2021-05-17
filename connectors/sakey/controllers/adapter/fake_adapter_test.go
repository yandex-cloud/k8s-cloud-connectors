// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
	"k8s-connectors/pkg/errors"
)

func setup() (context.Context, StaticAccessKeyAdapter) {
	ad := NewFakeStaticAccessKeyAdapter()
	return context.Background(), &ad
}

func createSakey(metaName, namespace, saID string) *connectorsv1.StaticAccessKey {
	return &connectorsv1.StaticAccessKey{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metaName,
			Namespace: namespace,
		},
		Spec: connectorsv1.StaticAccessKeySpec{
			ServiceAccountID: saID,
		},
	}
}

func TestRead(t *testing.T) {
	t.Run("read on one object returns one needed", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()
		sakey := createSakey("sakey", "default", "abdullah")
		skResp, err := ad.Create(ctx, sakey.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey.ClusterName, sakey.Name))
		require.NoError(t, err)

		// Act
		res, err := ad.Read(ctx, skResp.AccessKey.Id)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, skResp.AccessKey, res)
	})

	t.Run("read on multiple objects returns one needed", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()
		sakey1 := createSakey("sakey1", "default", "abdullah")
		sakey2 := createSakey("sakey2", "default", "sukhov")
		_, err := ad.Create(ctx, sakey1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey1.ClusterName, sakey1.Name))
		require.NoError(t, err)
		skResp, err := ad.Create(ctx, sakey2.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey2.ClusterName, sakey2.Name))
		require.NoError(t, err)

		// Act
		res, err := ad.Read(ctx, skResp.AccessKey.Id)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, skResp.AccessKey, res)
	})

	t.Run("read on no object throws", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()

		// Act
		_, err := ad.Read(ctx, "non-existent-id")

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.CheckRPCErrorNotFound(err))
	})
}

func TestList(t *testing.T) {
	t.Run("list on one object returns list of it", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()
		sakey := createSakey("sakey", "default", "abdullah")
		skResp, err := ad.Create(ctx, sakey.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey.ClusterName, sakey.Name))
		require.NoError(t, err)

		// Act
		res, err := ad.List(ctx, "abdullah")
		require.NoError(t, err)

		// Assert
		assert.Len(t, res, 1)
		assert.Contains(t, res, skResp.AccessKey)
	})

	t.Run("list on multiple objects returns list of them", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()
		sakey1 := createSakey("sakey1", "default", "abdullah")
		sakey2 := createSakey("sakey2", "default", "sukhov")
		sakey3 := createSakey("sakey3", "default", "sukhov")
		_, err := ad.Create(ctx, sakey1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey1.ClusterName, sakey1.Name))
		require.NoError(t, err)
		sk2Resp, err := ad.Create(ctx, sakey2.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey2.ClusterName, sakey2.Name))
		require.NoError(t, err)
		sk3Resp, err := ad.Create(ctx, sakey3.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey3.ClusterName, sakey3.Name))
		require.NoError(t, err)

		// Act
		res, err := ad.List(ctx, "sukhov")
		require.NoError(t, err)

		// Assert
		assert.Len(t, res, 2)
		assert.Contains(t, res, sk2Resp.AccessKey)
		assert.Contains(t, res, sk3Resp.AccessKey)
	})

	t.Run("list on no objects returns empty list", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()

		// Act
		res, err := ad.List(ctx, "gulchatay")
		require.NoError(t, err)

		// Assert
		assert.Len(t, res, 0)
	})
}

func TestDelete(t *testing.T) {
	t.Run("delete on this object deletes it", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()
		sakey := createSakey("sakey", "default", "abdullah")
		skResp, err := ad.Create(ctx, sakey.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey.ClusterName, sakey.Name))
		require.NoError(t, err)

		// Act
		err = ad.Delete(ctx, skResp.AccessKey.Id)
		require.NoError(t, err)
		res, err := ad.List(ctx, "abdullah")
		require.NoError(t, err)

		// Assert
		assert.Len(t, res, 0)
	})

	t.Run("delete on other objects throws", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()
		sakey1 := createSakey("sakey1", "default", "abdullah")
		sakey2 := createSakey("sakey2", "default", "sukhov")
		sk1Resp, err := ad.Create(ctx, sakey1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey1.ClusterName, sakey1.Name))
		require.NoError(t, err)
		sk2Resp, err := ad.Create(ctx, sakey2.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey2.ClusterName, sakey2.Name))
		require.NoError(t, err)

		// Act
		err1 := ad.Delete(ctx, "non-existent-id")
		res1, err := ad.List(ctx, "abdullah")
		require.NoError(t, err)
		res2, err := ad.List(ctx, "sukhov")
		require.NoError(t, err)

		// Assert
		assert.Error(t, err1)
		assert.Len(t, res1, 1)
		assert.Len(t, res2, 1)
		assert.Contains(t, res1, sk1Resp.AccessKey)
		assert.Contains(t, res2, sk2Resp.AccessKey)
	})

	t.Run("delete on this and other objects deletes this", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()
		sakey1 := createSakey("sakey1", "default", "abdullah")
		sakey2 := createSakey("sakey2", "default", "sukhov")
		sakey3 := createSakey("sakey3", "default", "sukhov")
		sk1Resp, err := ad.Create(ctx, sakey1.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey1.ClusterName, sakey1.Name))
		require.NoError(t, err)
		sk2Resp, err := ad.Create(ctx, sakey2.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey2.ClusterName, sakey2.Name))
		require.NoError(t, err)
		sk3Resp, err := ad.Create(ctx, sakey3.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(sakey3.ClusterName, sakey3.Name))
		require.NoError(t, err)

		// Act
		require.NoError(t, ad.Delete(ctx, sk2Resp.AccessKey.Id))
		res1, err := ad.List(ctx, "abdullah")
		require.NoError(t, err)
		res2, err := ad.List(ctx, "sukhov")
		require.NoError(t, err)

		// Assert
		assert.Len(t, res1, 1)
		assert.Len(t, res2, 1)
		assert.Contains(t, res1, sk1Resp.AccessKey)
		assert.Contains(t, res2, sk3Resp.AccessKey)
	})

	t.Run("delete on no object throws", func(t *testing.T) {
		// Arrange
		ctx, ad := setup()

		// Act
		err := ad.Delete(ctx, "non-existent-id")

		// Assert
		assert.Error(t, err)
	})
}
