// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s-connectors/pkg/config"
)

func TestAllocate(t *testing.T) {
	t.Run(
		"allocate on empty cloud creates resource", func(t *testing.T) {
			// Arrange
			ctx, log, _, ad, rc := setup(t)
			obj := createObject("registry", "folder", "obj", "default")

			// Act
			res, err := rc.allocateResource(ctx, log, &obj)
			require.NoError(t, err)
			lst, err := ad.List(ctx, "folder")
			require.NoError(t, err)

			// Assert
			assert.Len(t, lst, 1)
			// Check match with cloud
			assert.Equal(t, "registry", lst[0].Name)
			assert.Equal(t, "folder", lst[0].FolderId)
			assert.Equal(t, "test-cluster", lst[0].Labels[config.CloudClusterLabel])
			assert.Equal(t, "obj", lst[0].Labels[config.CloudNameLabel])
			// Check match with returned object
			assert.Equal(t, "registry", res.Name)
			assert.Equal(t, "folder", res.FolderId)
			assert.Equal(t, "test-cluster", res.Labels[config.CloudClusterLabel])
			assert.Equal(t, "obj", res.Labels[config.CloudNameLabel])
		},
	)

	t.Run(
		"allocate on non-empty cloud creates resource", func(t *testing.T) {
			// Arrange
			ctx, log, _, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")

			_ = createResourceRequireNoError(ctx, ad, t, "resource1", "folder", "obj1", "")
			_ = createResourceRequireNoError(ctx, ad, t, "resource2", "other-folder", "obj2", "")

			// Act
			_, err := rc.allocateResource(ctx, log, &obj)
			require.NoError(t, err)
			lst1, err := ad.List(ctx, "folder")
			require.NoError(t, err)
			lst2, err := ad.List(ctx, "other-folder")
			require.NoError(t, err)

			// Assert
			assert.Len(t, lst1, 2)
			assert.Len(t, lst2, 1)
		},
	)
}

func TestDeallocate(t *testing.T) {
	t.Run(
		"deallocate on cloud with resource deletes resource", func(t *testing.T) {
			// Arrange
			ctx, log, _, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")
			_, err := rc.allocateResource(ctx, log, &obj)
			require.NoError(t, err)

			// Act
			require.NoError(t, rc.deallocateResource(ctx, log, &obj))
			lst, err := ad.List(ctx, "folder")
			require.NoError(t, err)

			// Assert
			assert.Len(t, lst, 0)
		},
	)

	t.Run(
		"deallocate on cloud with this and other resources deletes this resource", func(t *testing.T) {
			// Arrange
			ctx, log, _, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")

			otherObj1 := createObject("other-resource", "folder", "otherObj1", "default")
			otherObj2 := createObject("resource", "other-folder", "otherObj2", "default")

			_, err := rc.allocateResource(ctx, log, &obj)
			require.NoError(t, err)
			_, err = rc.allocateResource(ctx, log, &otherObj1)
			require.NoError(t, err)
			_, err = rc.allocateResource(ctx, log, &otherObj2)
			require.NoError(t, err)

			// Act
			require.NoError(t, rc.deallocateResource(ctx, log, &obj))
			lst1, err := ad.List(ctx, "folder")
			require.NoError(t, err)
			lst2, err := ad.List(ctx, "other-folder")
			require.NoError(t, err)

			// Assert
			assert.Len(t, lst1, 1)
			assert.Equal(t, "other-resource", lst1[0].Name)
			assert.Equal(t, "folder", lst1[0].FolderId)
			assert.Equal(t, "test-cluster", lst1[0].Labels[config.CloudClusterLabel])
			assert.Equal(t, "otherObj1", lst1[0].Labels[config.CloudNameLabel])
			assert.Len(t, lst2, 1)
		},
	)

	t.Run(
		"deallocate on cloud without resource does nothing", func(t *testing.T) {
			// Arrange
			ctx, log, _, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")

			// Act
			require.NoError(t, rc.deallocateResource(ctx, log, &obj))
			lst, err := ad.List(ctx, "folder")
			require.NoError(t, err)

			// Assert
			assert.Len(t, lst, 0)
		},
	)
}
