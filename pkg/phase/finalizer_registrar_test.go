// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/yandex-cloud/k8s-cloud-connectors/pkg/util"
)

func TestRegisterFinalizer(t *testing.T) {
	t.Run(
		"register on empty finalizer list adds finalizer", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			object := v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "resource",
					Namespace:  "default",
					Finalizers: []string{},
				},
			}
			require.NoError(t, cl.Create(ctx, &object))

			// Act
			require.NoError(t, RegisterFinalizer(ctx, cl, log, &object.ObjectMeta, &object, "good"))
			var res v1.Pod
			require.NoError(t, cl.Get(ctx, util.NamespacedName(&object), &res))

			// Assert
			assert.Len(t, res.Finalizers, 1)
			assert.Contains(t, res.Finalizers, "good")
		},
	)

	t.Run(
		"update on non-empty finalizer list adds finalizer", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			object := v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "object",
					Namespace:  "default",
					Finalizers: []string{"bad", "ugly"},
				},
			}
			require.NoError(t, cl.Create(ctx, &object))

			// Act
			require.NoError(t, RegisterFinalizer(ctx, cl, log, &object.ObjectMeta, &object, "good"))
			var res v1.Pod
			require.NoError(t, cl.Get(ctx, util.NamespacedName(&object), &res))

			// Assert
			assert.Len(t, res.Finalizers, 3)
			assert.Contains(t, res.Finalizers, "good")
			assert.Contains(t, res.Finalizers, "bad")
			assert.Contains(t, res.Finalizers, "ugly")
		},
	)
}

func TestDeregisterFinalizer(t *testing.T) {
	t.Run(
		"deregister on empty finalizer list does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			object := v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "resource",
					Namespace:  "default",
					Finalizers: []string{},
				},
			}
			require.NoError(t, cl.Create(ctx, &object))

			// Act
			require.NoError(t, DeregisterFinalizer(ctx, cl, log, &object.ObjectMeta, &object, "good"))
			var res v1.Pod
			require.NoError(t, cl.Get(ctx, util.NamespacedName(&object), &res))

			// Assert
			assert.Len(t, res.Finalizers, 0)
		},
	)

	t.Run(
		"deregister on finalizer with other finalizers list does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			object := v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "resource",
					Namespace:  "default",
					Finalizers: []string{"bad", "ugly"},
				},
			}
			require.NoError(t, cl.Create(ctx, &object))

			// Act
			require.NoError(t, DeregisterFinalizer(ctx, cl, log, &object.ObjectMeta, &object, "good"))
			var res v1.Pod
			require.NoError(t, cl.Get(ctx, util.NamespacedName(&object), &res))

			// Assert
			assert.Len(t, res.Finalizers, 2)
			assert.Contains(t, res.Finalizers, "bad")
			assert.Contains(t, res.Finalizers, "ugly")
		},
	)

	t.Run(
		"deregister on non-empty finalizer list removes finalizer", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			object := v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "resource",
					Namespace:  "default",
					Finalizers: []string{"good", "bad", "ugly"},
				},
			}
			require.NoError(t, cl.Create(ctx, &object))
			// Act
			require.NoError(t, DeregisterFinalizer(ctx, cl, log, &object.ObjectMeta, &object, "good"))
			var res v1.Pod
			require.NoError(t, cl.Get(ctx, util.NamespacedName(&object), &res))

			// Assert
			assert.Len(t, res.Finalizers, 2)
			assert.Contains(t, res.Finalizers, "bad")
			assert.Contains(t, res.Finalizers, "ugly")
		},
	)

	t.Run(
		"deregister on finalizer list with only this finalizer removes finalizer", func(t *testing.T) {
			// Arrange
			ctx, log, cl := setup(t)
			object := v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "resource",
					Namespace:  "default",
					Finalizers: []string{"good"},
				},
			}
			require.NoError(t, cl.Create(ctx, &object))
			// Act
			require.NoError(t, DeregisterFinalizer(ctx, cl, log, &object.ObjectMeta, &object, "good"))
			var res v1.Pod
			require.NoError(t, cl.Get(ctx, util.NamespacedName(&object), &res))

			// Assert
			assert.Len(t, res.Finalizers, 0)
		},
	)
}
