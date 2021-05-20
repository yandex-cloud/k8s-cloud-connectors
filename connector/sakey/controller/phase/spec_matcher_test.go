// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s-connectors/connector/sakey/controller/adapter"
	sakeyconfig "k8s-connectors/connector/sakey/pkg/config"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setupSpecMatcher(t *testing.T) (
	context.Context, logr.Logger, client.Client, adapter.StaticAccessKeyAdapter, StaticAccessKeyPhase,
) {
	ad := adapter.NewFakeStaticAccessKeyAdapter()
	cl := k8sfake.NewFakeClient()
	return context.Background(), logrfake.NewFakeLogger(t), cl, &ad, &SpecMatcher{
		Sdk: &ad,
	}
}

func TestSpecMatcherIsUpdated(t *testing.T) {
	t.Run(
		"is updated on matching spec", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, phase := setupSpecMatcher(t)
			obj := createObject("sukhov", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))
			_, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj.ClusterName, obj.Name),
			)
			require.NoError(t, err)

			// Act
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.True(t, upd)
		},
	)

	t.Run(
		"is not updated on non-matching spec", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, phase := setupSpecMatcher(t)
			obj := createObject("sukhov", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))
			resp, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj.ClusterName, obj.Name),
			)
			require.NoError(t, err)
			obj.Status.KeyID = resp.AccessKey.Id
			obj.Spec.ServiceAccountID = "abdullah"

			// Act
			upd, err := phase.IsUpdated(ctx, log, &obj)
			require.NoError(t, err)

			// Assert
			assert.False(t, upd)
		},
	)
}

func TestSpecMatcherUpdate(t *testing.T) {
	t.Run(
		"update always fails", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, phase := setupSpecMatcher(t)
			obj := createObject("sukhov", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))
			_, err := ad.Create(
				ctx, obj.Spec.ServiceAccountID, sakeyconfig.GetStaticAccessKeyDescription(obj.ClusterName, obj.Name),
			)
			require.NoError(t, err)

			// Act
			err = phase.Update(ctx, log, &obj)

			// Assert
			assert.Error(t, err)
		},
	)
}
