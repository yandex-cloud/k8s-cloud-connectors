// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"

	v1 "k8s-connectors/connector/yos/api/v1"
	"k8s-connectors/pkg/webhook"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupValidation(t *testing.T) (context.Context, webhook.Validator, logr.Logger) {
	t.Helper()
	return context.TODO(), &YOSValidator{}, logrfake.NewFakeLogger(t)
}

func TestUpdateValidation(t *testing.T) {
	t.Run("no change is valid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log := setupValidation(t)
		old := v1.YandexObjectStorage{
			Spec: v1.YandexObjectStorageSpec{
				Name:      "yos",
				ACL:       "",
				SAKeyName: "",
			},
		}
		current := v1.YandexObjectStorage{
			Spec: v1.YandexObjectStorageSpec{
				Name:      "yos",
				ACL:       "",
				SAKeyName: "",
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
		old := v1.YandexObjectStorage{
			Spec: v1.YandexObjectStorageSpec{
				Name:      "yos",
				ACL:       "",
				SAKeyName: "",
			},
		}
		current := v1.YandexObjectStorage{
			Spec: v1.YandexObjectStorageSpec{
				Name:      "other-yos",
				ACL:       "",
				SAKeyName: "",
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, &webhook.ValidationError{}))
	})
}
