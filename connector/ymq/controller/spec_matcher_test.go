// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ymqutil "k8s-connectors/connector/ymq/pkg/util"
)

func TestSpecMatching(t *testing.T) {
	t.Run("match spec on equal specs does nothing", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj := createDefaultQueue("obj", "default", "sakey", "queue")
		require.NoError(t, cl.Create(ctx, &obj))
		require.NoError(t, rc.allocateResource(ctx, log, &obj, nil))

		// Act
		require.NoError(t, rc.matchSpec(ctx, log, &obj, nil))
		res, err := ad.GetAttributes(ctx, nil, obj.Status.QueueURL)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, "0", *res[ymqutil.DelaySeconds])
		assert.Equal(t, "262144", *res[ymqutil.MaximumMessageSize])
		assert.Equal(t, "60", *res[ymqutil.MessageRetentionPeriod])
		assert.Equal(t, "0", *res[ymqutil.ReceiveMessageWaitTimeSeconds])
		assert.Equal(t, "30", *res[ymqutil.VisibilityTimeout])
	})

	t.Run("match spec on non-equal specs matches cloud resource with object", func(t *testing.T) {
		// Arrange
		ctx, log, cl, ad, rc := setup(t)
		createSAKeyRequireNoError(ctx, t, cl, "sakey", "default")
		obj := createDefaultQueue("obj", "default", "sakey", "queue")
		require.NoError(t, cl.Create(ctx, &obj))
		require.NoError(t, rc.allocateResource(ctx, log, &obj, nil))

		// Act
		obj.Spec.DelaySeconds = 1
		obj.Spec.MaximumMessageSize = 262143
		obj.Spec.MessageRetentionPeriod = 59
		obj.Spec.ReceiveMessageWaitTimeSeconds = 1
		obj.Spec.VisibilityTimeout = 29
		require.NoError(t, rc.matchSpec(ctx, log, &obj, nil))
		res, err := ad.GetAttributes(ctx, nil, obj.Status.QueueURL)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, "1", *res[ymqutil.DelaySeconds])
		assert.Equal(t, "262143", *res[ymqutil.MaximumMessageSize])
		assert.Equal(t, "59", *res[ymqutil.MessageRetentionPeriod])
		assert.Equal(t, "1", *res[ymqutil.ReceiveMessageWaitTimeSeconds])
		assert.Equal(t, "29", *res[ymqutil.VisibilityTimeout])
	})
}
