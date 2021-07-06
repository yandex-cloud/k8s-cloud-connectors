// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"

	v1 "k8s-connectors/connector/ymq/api/v1"
	"k8s-connectors/pkg/webhook"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupValidation(t *testing.T) (context.Context, webhook.Validator, logr.Logger) {
	t.Helper()
	return context.TODO(), &YMQValidator{}, logrfake.NewFakeLogger(t)
}

func TestCreateValidation(t *testing.T) {
	t.Run("usual queue without fifo suffix is valid", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		obj := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:      "q",
				FifoQueue: false,
			},
		}

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("fifo queue with fifo suffix is valid", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		obj := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:      "q.fifo",
				FifoQueue: true,
			},
		}

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("fifo queue without fifo suffix is invalid", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		obj := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:      "q",
				FifoQueue: true,
			},
		}

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, webhook.ValidationError{}))
	})

	t.Run("usual queue with fifo suffix is invalid", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		obj := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:      "q.fifo",
				FifoQueue: false,
			},
		}

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, webhook.ValidationError{}))
	})

	t.Run("usual queue with content based deduplication is invalid", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		obj := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:                      "q",
				FifoQueue:                 false,
				ContentBasedDeduplication: true,
			},
		}

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, webhook.ValidationError{}))
	})
}

func TestUpdateValidation(t *testing.T) {
	t.Run("no change is valid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		old := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:                      "q",
				FifoQueue:                 false,
				ContentBasedDeduplication: true,
			},
		}
		current := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:                      "q",
				FifoQueue:                 false,
				ContentBasedDeduplication: true,
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("name change is invalid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		old := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name: "q",
			},
		}
		current := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name: "queue",
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, webhook.ValidationError{}))
	})

	t.Run("queue type change is invalid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		old := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:      "q.fifo",
				FifoQueue: true,
			},
		}
		current := v1.YandexMessageQueue{
			Spec: v1.YandexMessageQueueSpec{
				Name:      "q",
				FifoQueue: false,
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, webhook.ValidationError{}))
	})
}
