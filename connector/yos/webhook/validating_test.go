// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package webhook

import (
	"context"
	"errors"
	"testing"

	sakey "k8s-connectors/connector/sakey/api/v1"
	k8sfake "k8s-connectors/testing/k8s-fake"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s-connectors/connector/yos/api/v1"
	"k8s-connectors/pkg/webhook"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupValidation(t *testing.T) (context.Context, webhook.Validator, logr.Logger, client.Client) {
	t.Helper()
	cl := k8sfake.NewFakeClient()
	return context.TODO(), &YOSValidator{cl: cl}, logrfake.NewFakeLogger(t), cl
}

func createSAKey(ctx context.Context, t *testing.T, cl client.Client, name, namespace string) {
	t.Helper()
	require.NoError(
		t, cl.Create(
			ctx, &sakey.StaticAccessKey{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			},
		),
	)
}

func TestCreateValidation(t *testing.T) {
	t.Run("create on an existent SAKey is valid", func(t *testing.T) {
		// Arrange
		ctx, wh, log, cl := setupValidation(t)
		obj := v1.YandexObjectStorage{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "some-namespace",
			},
			Spec: v1.YandexObjectStorageSpec{
				Name:      "bucket",
				SAKeyName: "real-sakey",
			},
		}
		createSAKey(ctx, t, cl, "real-sakey", "some-namespace")

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("create on a non-existent SAKey is invalid", func(t *testing.T) {
		// Arrange
		ctx, wh, log, _ := setupValidation(t)
		obj := v1.YandexObjectStorage{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "some-namespace",
			},
			Spec: v1.YandexObjectStorageSpec{
				Name:      "bucket",
				SAKeyName: "real-sakey",
			},
		}

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, &webhook.ValidationError{}))
	})

	t.Run("create on another SAKey is invalid", func(t *testing.T) {
		// Arrange
		ctx, wh, log, cl := setupValidation(t)
		obj := v1.YandexObjectStorage{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "some-namespace",
			},
			Spec: v1.YandexObjectStorageSpec{
				Name:      "bucket",
				SAKeyName: "real-sakey",
			},
		}
		createSAKey(ctx, t, cl, "another-sakey", "some-namespace")

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, &webhook.ValidationError{}))
	})

	t.Run("create on SAKey in another namespace is invalid", func(t *testing.T) {
		// Arrange
		ctx, wh, log, cl := setupValidation(t)
		obj := v1.YandexObjectStorage{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "some-namespace",
			},
			Spec: v1.YandexObjectStorageSpec{
				Name:      "bucket",
				SAKeyName: "real-sakey",
			},
		}
		createSAKey(ctx, t, cl, "real-sakey", "default")

		// Act
		err := wh.ValidateCreation(ctx, log, &obj)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, &webhook.ValidationError{}))
	})
}

func TestUpdateValidation(t *testing.T) {
	t.Run("no change is valid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log, _ := setupValidation(t)
		old := v1.YandexObjectStorage{
			Spec: v1.YandexObjectStorageSpec{
				Name: "yos",
			},
		}
		current := v1.YandexObjectStorage{
			Spec: v1.YandexObjectStorageSpec{
				Name: "yos",
			},
		}

		// Act
		err := wh.ValidateUpdate(ctx, log, &current, &old)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("name change is invalid update", func(t *testing.T) {
		// Arrange
		ctx, wh, log, _ := setupValidation(t)
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

func TestDeleteValidate(t *testing.T) {
	// TODO: test with aws sdk mock
}
