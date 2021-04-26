// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
	sakeyconfig "k8s-connectors/connectors/sakey/pkg/config"
	"k8s-connectors/pkg/utils"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestIsUpdated(t *testing.T) {
	t.Run("empty finalizers means not updated", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		updated, err := phase.IsUpdated(context.Background(), log, &resource)
		require.NoError(t, err)

		// Assert
		assert.False(t, updated)
	})

	t.Run("other finalizers means not updated", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{"not.that.finalizer", "yet.another.false.finalizer"},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		updated, err := phase.IsUpdated(context.Background(), log, &resource)
		require.NoError(t, err)

		// Assert
		assert.False(t, updated)
	})

	t.Run("finalizer exist means updated", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{sakeyconfig.FinalizerName},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		updated, err := phase.IsUpdated(context.Background(), log, &resource)
		require.NoError(t, err)

		// Assert
		assert.True(t, updated)
	})

	t.Run("finalizer and others exist means updated", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{"not.that.finalizer", sakeyconfig.FinalizerName, "yet.another.false.finalizer"},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		updated, err := phase.IsUpdated(context.Background(), log, &resource)
		require.NoError(t, err)

		// Assert
		assert.True(t, updated)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("update on empty finalizer list adds finalizer", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		require.NoError(t, phase.Update(context.Background(), log, &resource))

		// Assert
		var res connectorsv1.StaticAccessKey
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Len(t, res.Finalizers, 1)
		assert.Contains(t, res.Finalizers, sakeyconfig.FinalizerName)
	})

	t.Run("update on non-empty finalizer list adds finalizer", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{"not.that.finalizer", "yet.another.finalizer"},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		require.NoError(t, phase.Update(context.Background(), log, &resource))

		// Assert
		var res connectorsv1.StaticAccessKey
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Len(t, res.Finalizers, 3)
		assert.Contains(t, res.Finalizers, "not.that.finalizer")
		assert.Contains(t, res.Finalizers, "yet.another.finalizer")
		assert.Contains(t, res.Finalizers, sakeyconfig.FinalizerName)
	})
}

func TestCleanup(t *testing.T) {
	t.Run("cleanup on empty finalizer list does nothing", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		require.NoError(t, phase.Cleanup(context.Background(), log, &resource))

		// Assert
		var res connectorsv1.StaticAccessKey
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Len(t, res.Finalizers, 0)
	})

	t.Run("cleanup on non-empty finalizer list removes finalizer", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{"not.that.finalizer", sakeyconfig.FinalizerName, "yet.another.finalizer"},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		require.NoError(t, phase.Cleanup(context.Background(), log, &resource))

		// Assert
		var res connectorsv1.StaticAccessKey
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Len(t, res.Finalizers, 2)
		assert.Contains(t, res.Finalizers, "not.that.finalizer")
		assert.Contains(t, res.Finalizers, "yet.another.finalizer")
	})

	t.Run("cleanup on non-empty finalizer list without needed finalizer does nothing", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.StaticAccessKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{"not.that.finalizer", "yet.another.finalizer"},
			},
		}
		c := k8sfake.NewFakeClient()
		log := logrfake.NewFakeLogger(t)
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		require.NoError(t, phase.Cleanup(context.Background(), log, &resource))

		// Assert
		var res connectorsv1.StaticAccessKey
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Len(t, res.Finalizers, 2)
		assert.Contains(t, res.Finalizers, "not.that.finalizer")
		assert.Contains(t, res.Finalizers, "yet.another.finalizer")
	})
}
