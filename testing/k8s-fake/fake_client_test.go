// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package k8s_fake

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestCreate(t *testing.T) {
	// Arrange
	c := NewFakeClient()
	ctx := context.Background()

	// Act
	err1 := c.Create(ctx, &v1.Secret{
		TypeMeta:   metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "secret",
			Namespace:                  "default",
			Generation:                 0,
		},
		Immutable:  nil,
		Data:       nil,
		StringData: map[string]string{
			"secret-specific-data" : "exists",
		},
		Type:       "opaque",
	})

	var res v1.Secret
	err2 := c.Get(ctx, client.ObjectKey{
		Name: "secret",
		Namespace: "default",
	}, &res)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "secret", res.Name)
	assert.Equal(t, "default", res.Namespace)
	assert.Equal(t, "exists", res.StringData["secret-specific-data"])
}

func TestCreateDelete(t *testing.T) {
	// Arrange
	c := NewFakeClient()
	ctx := context.Background()

	// Act
	secret := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       "secret",
			Namespace:                  "default",
			Generation:                 0,
		},
		Immutable:  nil,
		Data:       nil,
		StringData: map[string]string{
			"secret-specific-data" : "exists",
		},
		Type:       "opaque",
	}

	err1 := c.Create(ctx, secret)
	err2 := c.Delete(ctx, secret)

	var res v1.Secret
	err3 := c.Get(ctx, client.ObjectKey{
		Name: "secret",
		Namespace: "default",
	}, &res)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Error(t, err3)
	assert.True(t, errors.IsNotFound(err3))
}
