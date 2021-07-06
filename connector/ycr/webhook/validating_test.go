// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"

	v1 "k8s-connectors/connector/ycr/api/v1"
	"k8s-connectors/pkg/webhook"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupValidation(t *testing.T) (context.Context, webhook.Validator, logr.Logger) {
	t.Helper()
	return context.TODO(), &YCRValidator{}, logrfake.NewFakeLogger(t)
}

func TestUpdateValidation(t *testing.T) {
	t.Run("name change is valid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		old := v1.YandexContainerRegistry{
			Spec: v1.YandexContainerRegistrySpec{
				Name:     "res",
				FolderID: "folder",
			},
		}
		current := v1.YandexContainerRegistry{
			Spec: v1.YandexContainerRegistrySpec{
				Name:     "other-res",
				FolderID: "folder",
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("no change is valid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		old := v1.YandexContainerRegistry{
			Spec: v1.YandexContainerRegistrySpec{
				Name:     "res",
				FolderID: "folder",
			},
		}
		current := v1.YandexContainerRegistry{
			Spec: v1.YandexContainerRegistrySpec{
				Name:     "res",
				FolderID: "folder",
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("folder change is invalid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		old := v1.YandexContainerRegistry{
			Spec: v1.YandexContainerRegistrySpec{
				Name:     "res",
				FolderID: "folder",
			},
		}
		current := v1.YandexContainerRegistry{
			Spec: v1.YandexContainerRegistrySpec{
				Name:     "res",
				FolderID: "other-folder",
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, webhook.ValidationError{}))
	})

	t.Run("folder and name change is invalid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		old := v1.YandexContainerRegistry{
			Spec: v1.YandexContainerRegistrySpec{
				Name:     "res",
				FolderID: "folder",
			},
		}
		current := v1.YandexContainerRegistry{
			Spec: v1.YandexContainerRegistrySpec{
				Name:     "other-res",
				FolderID: "other-folder",
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, webhook.ValidationError{}))
	})
}
