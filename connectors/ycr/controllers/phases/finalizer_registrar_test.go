// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	ycrconfig "k8s-connectors/connectors/ycr/pkg/config"
	"k8s-connectors/pkg/utils"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestIsUpdated(t *testing.T) {
	t.Run("empty finalizers means not updated", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{},
			},
		}
		c := k8sfake.NewFakeClient()
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		updated, err := phase.IsUpdated(context.Background(), &resource)
		require.NoError(t, err)

		// Assert
		assert.False(t, updated)
	})

	t.Run("other finalizers means not updated", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{"not.that.finalizer", "yet.another.false.finalizer"},
			},
		}
		c := k8sfake.NewFakeClient()
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		updated, err := phase.IsUpdated(context.Background(), &resource)
		require.NoError(t, err)

		// Assert
		assert.False(t, updated)
	})

	t.Run("finalizer exist means updated", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{ycrconfig.FinalizerName},
			},
		}
		c := k8sfake.NewFakeClient()
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		updated, err := phase.IsUpdated(context.Background(), &resource)
		require.NoError(t, err)

		// Assert
		assert.True(t, updated)
	})

	t.Run("finalizer and others exist means updated", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{"not.that.finalizer", ycrconfig.FinalizerName, "yet.another.false.finalizer"},
			},
		}
		c := k8sfake.NewFakeClient()
		phase := FinalizerRegistrar{
			Client: &c,
		}
		require.NoError(t, c.Create(context.Background(), &resource))

		// Act
		updated, err := phase.IsUpdated(context.Background(), &resource)
		require.NoError(t, err)

		// Assert
		assert.True(t, updated)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("update on empty finalizer list adds finalizer", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
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
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Equal(t, []string{ycrconfig.FinalizerName}, res.Finalizers)
	})

	t.Run("update on non-empty finalizer list adds finalizer", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
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
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Equal(t, []string{"not.that.finalizer", "yet.another.finalizer", ycrconfig.FinalizerName}, res.Finalizers)
	})
}

func TestCleanup(t *testing.T) {
	t.Run("cleanup on empty finalizer list does nothing", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
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
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Equal(t, []string{}, res.Finalizers)
	})

	t.Run("cleanup on non-empty finalizer list removes finalizer", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "resource",
				Namespace:  "default",
				Finalizers: []string{"not.that.finalizer", ycrconfig.FinalizerName, "yet.another.finalizer"},
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
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Equal(t, []string{"not.that.finalizer", "yet.another.finalizer"}, res.Finalizers)
	})

	t.Run("cleanup on non-empty finalizer list without needed finalizer does nothing", func(t *testing.T) {
		// Arrange
		resource := connectorsv1.YandexContainerRegistry{
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
		var res connectorsv1.YandexContainerRegistry
		require.NoError(t, c.Get(context.Background(), utils.NamespacedName(&resource), &res))
		assert.Equal(t, []string{"not.that.finalizer", "yet.another.finalizer"}, res.Finalizers)
	})
}
