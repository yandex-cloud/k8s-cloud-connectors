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
		obj := createDefaultQueue("obj", "default", "sakey", "queue")
		require.NoError(t, cl.Create(ctx, &obj))

		// Act
		require.NoError(t, rc.allocateResource(ctx, log, &obj, "test-key", "test-secret"))
		lst, err := ad.List(ctx, "test-key", "test-secret")
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 1)
	})

	t.Run("allocate on non-empty cloud allocates resource", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj1 := createDefaultQueue("obj1", "default", "sakey", "first-queue")
		require.NoError(t, cl.Create(ctx, &obj1))
		require.NoError(t, rc.allocateResource(ctx, log, &obj1, "test-key", "test-secret"))

		obj2 := createDefaultQueue("obj2", "default", "sakey", "second-queue")
		require.NoError(t, cl.Create(ctx, &obj2))

		// Act
		require.NoError(t, rc.allocateResource(ctx, log, &obj2, "test-key", "test-secret"))
		lst, err := ad.List(ctx, "test-key", "test-secret")
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 2)
	})

	t.Run("allocate on cloud with queue with same name and attributes does nothing", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj1 := createDefaultQueue("obj1", "default", "sakey", "queue")
		require.NoError(t, cl.Create(ctx, &obj1))
		require.NoError(t, rc.allocateResource(ctx, log, &obj1, "test-key", "test-secret"))

		obj2 := createDefaultQueue("obj2", "default", "sakey", "queue")
		require.NoError(t, cl.Create(ctx, &obj2))

		// Act
		require.NoError(t, rc.allocateResource(ctx, log, &obj2, "test-key", "test-secret"))
		lst, err := ad.List(ctx, "test-key", "test-secret")
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 1)
	})

	t.Run("allocate on cloud with queue with same name and different attributes fails", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj1 := createDefaultQueue("obj1", "default", "sakey", "queue")
		require.NoError(t, cl.Create(ctx, &obj1))
		require.NoError(t, rc.allocateResource(ctx, log, &obj1, "test-key", "test-secret"))

		obj2 := createDefaultQueue("obj2", "default", "sakey", "queue")
		obj2.Spec.DelaySeconds = 1
		require.NoError(t, cl.Create(ctx, &obj2))

		// Act
		err1 := rc.allocateResource(ctx, log, &obj2, "test-key", "test-secret")
		lst, err := ad.List(ctx, "test-key", "test-secret")
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
		obj := createDefaultQueue("obj", "default", "sakey", "queue")
		// Some random URL so it is not malformed
		obj.Status.QueueURL = "https://message-queue.api.cloud.yandex.net/queue/url/sqs"
		require.NoError(t, cl.Create(ctx, &obj))

		// Act
		require.NoError(t, rc.deallocateResource(ctx, log, &obj, "test-key", "test-secret"))
		lst, err := ad.List(ctx, "test-key", "test-secret")
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 0)
	})

	t.Run("deallocate on cloud with this resource deletes this resource", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj := createDefaultQueue("obj", "default", "sakey", "queue")
		require.NoError(t, cl.Create(ctx, &obj))
		require.NoError(t, rc.allocateResource(ctx, log, &obj, "test-key", "test-secret"))

		// Act
		require.NoError(t, rc.deallocateResource(ctx, log, &obj, "test-key", "test-secret"))
		lst, err := ad.List(ctx, "test-key", "test-secret")
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 0)
	})
}
