// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package k8sfake

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

func TestCreate(t *testing.T) {
	// Arrange
	c := NewFakeClient()
	ctx := context.Background()
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "secret",
			Namespace:  "default",
			Generation: 0,
		},
		Data: map[string][]byte{
			"secret-specific-data": []byte("exists"),
		},
		Type: "opaque",
	}

	// Act
	require.NoError(t, c.Create(ctx, secret))

	var res v1.Secret
	require.NoError(
		t, c.Get(
			ctx, client.ObjectKey{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			}, &res,
		),
	)

	// Assert
	assert.Equal(t, *secret, res)
}

func TestCreateDelete(t *testing.T) {
	// Arrange
	c := NewFakeClient()
	ctx := context.Background()
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "secret",
			Namespace:  "default",
			Generation: 0,
		},
		Immutable: nil,
		Data: map[string][]byte{
			"secret-specific-data": []byte("exists"),
		},
		Type: "opaque",
	}

	// Act
	require.NoError(t, c.Create(ctx, secret))
	require.NoError(t, c.Delete(ctx, secret))

	var res v1.Secret
	err := c.Get(
		ctx, client.ObjectKey{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		}, &res,
	)

	// Assert
	assert.True(t, errors.IsNotFound(err))
}

func TestUpdate(t *testing.T) {
	// Arrange
	c := NewFakeClient()
	ctx := context.Background()
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "secret",
			Namespace:  "default",
			Generation: 0,
		},
		Immutable: nil,
		Data: map[string][]byte{
			"secret-specific-data": []byte("exists"),
		},
		Type: "opaque",
	}

	updSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "secret",
			Namespace:  "default",
			Generation: 0,
		},
		Immutable: nil,
		Data: map[string][]byte{
			"secret-specific-data": []byte("does-not-exists"),
		},
		Type: "opaque",
	}

	// Act
	require.NoError(t, c.Create(ctx, secret))
	require.NoError(t, c.Update(ctx, updSecret))

	var res v1.Secret
	require.NoError(
		t, c.Get(
			ctx, client.ObjectKey{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			}, &res,
		),
	)

	// Assert
	assert.Equal(t, *updSecret, res)
}
