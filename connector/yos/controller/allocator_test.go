// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllocate(t *testing.T) {
	t.Run("allocate on empty cloud allocates resource", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj := createObject("bucket", "sakey", "", "obj", "default")
		require.NoError(t, cl.Create(ctx, &obj))

		// Act
		require.NoError(t, rc.allocateResource(ctx, log, &obj, nil))
		lst, err := ad.List(ctx, nil)
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 1)
		assert.Equal(t, "bucket", *lst[0].Name)
	})

	t.Run("allocate on non-empty cloud allocates resource", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj1 := createObject("first-bucket", "sakey", "", "obj1", "default")
		require.NoError(t, cl.Create(ctx, &obj1))
		require.NoError(t, rc.allocateResource(ctx, log, &obj1, nil))

		obj2 := createObject("second-bucket", "sakey", "", "obj2", "default")
		require.NoError(t, cl.Create(ctx, &obj2))

		// Act
		require.NoError(t, rc.allocateResource(ctx, log, &obj2, nil))
		lst, err := ad.List(ctx, nil)
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 2)
	})

	t.Run("allocate on cloud with bucket with same name fails", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj1 := createObject("first-bucket", "sakey", "", "obj1", "default")
		require.NoError(t, cl.Create(ctx, &obj1))
		require.NoError(t, rc.allocateResource(ctx, log, &obj1, nil))

		obj2 := createObject("first-bucket", "sakey", "", "obj2", "default")
		require.NoError(t, cl.Create(ctx, &obj2))

		// Act
		err1 := rc.allocateResource(ctx, log, &obj2, nil)
		lst, err := ad.List(ctx, nil)
		require.NoError(t, err)

		// Assert
		assert.Error(t, err1)
		assert.Len(t, lst, 1)
	})
}

func TestDeallocate(t *testing.T) {
	t.Run("deallocate on empty cloud does nothing", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj := createObject("bucket", "sakey", "", "obj", "default")
		require.NoError(t, cl.Create(ctx, &obj))

		// Act
		require.NoError(t, rc.deallocateResource(ctx, log, &obj, nil))
		lst, err := ad.List(ctx, nil)
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 0)
	})

	t.Run("deallocate on cloud with this resource deletes this resource", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj := createObject("bucket", "sakey", "", "obj", "default")
		require.NoError(t, cl.Create(ctx, &obj))
		require.NoError(t, rc.allocateResource(ctx, log, &obj, nil))

		// Act
		require.NoError(t, rc.deallocateResource(ctx, log, &obj, nil))
		lst, err := ad.List(ctx, nil)
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 0)
	})
}
