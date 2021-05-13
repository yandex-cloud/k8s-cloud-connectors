// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package adapter

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"testing"
)

func TestRead(t *testing.T) {
	t.Run("read on one object", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		res, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		reg, err := ad.Read(ctx, res.Id)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, reg1.FolderId, reg.FolderId)
		assert.Equal(t, reg1.Name, reg.Name)
		assert.Equal(t, reg1.Labels, reg.Labels)
	})

	t.Run("read on multiple objects", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		reg2 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg2",
			Labels:   map[string]string{"key": "label"},
		}
		reg3 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg3",
			Labels:   map[string]string{"key": "label"},
		}
		_, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)
		res, err := ad.Create(ctx, &reg2)
		require.NoError(t, err)
		_, err = ad.Create(ctx, &reg3)
		require.NoError(t, err)

		// Act
		reg, err := ad.Read(ctx, res.Id)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, reg2.FolderId, reg.FolderId)
		assert.Equal(t, reg2.Name, reg.Name)
		assert.Equal(t, reg2.Labels, reg.Labels)
	})

	t.Run("read on no object", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()

		// Act
		_, err := ad.Read(ctx, "non-existent-id")

		// Assert
		assert.Error(t, err)
	})
}

func TestList(t *testing.T) {
	t.Run("list on one object", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		_, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		lst, err := ad.List(ctx, "folder")
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 1)
		assert.Equal(t, reg1.FolderId, lst[0].FolderId)
		assert.Equal(t, reg1.Name, lst[0].Name)
		assert.Equal(t, reg1.Labels, lst[0].Labels)
	})

	t.Run("list on multiple objects", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		reg2 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg2",
			Labels:   map[string]string{"key": "label"},
		}
		reg3 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg3",
			Labels:   map[string]string{"key": "label"},
		}
		_, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)
		_, err = ad.Create(ctx, &reg2)
		require.NoError(t, err)
		_, err = ad.Create(ctx, &reg3)
		require.NoError(t, err)

		// Act
		lst, err := ad.List(ctx, "folder")
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 3)
	})

	t.Run("list on no object", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()

		// Act
		lst, err := ad.List(ctx, "folder")
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 0)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("update on one object on name field", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		reg, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		require.NoError(t, ad.Update(ctx, &containerregistry.UpdateRegistryRequest{
			RegistryId: reg.Id,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			Name:       "reg1_updated",
			Labels:     map[string]string{},
		}))
		res, err := ad.Read(ctx, reg.Id)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, reg1.FolderId, res.FolderId)
		assert.Equal(t, "reg1_updated", res.Name)
		assert.Equal(t, reg1.Labels, res.Labels)
	})

	t.Run("update on one object on labels field", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		reg, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		require.NoError(t, ad.Update(ctx, &containerregistry.UpdateRegistryRequest{
			RegistryId: reg.Id,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"labels"}},
			Name:       "reg1_updated",
			Labels:     map[string]string{},
		}))
		res, err := ad.Read(ctx, reg.Id)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, reg1.FolderId, res.FolderId)
		assert.Equal(t, reg1.Name, res.Name)
		assert.Equal(t, map[string]string{}, res.Labels)
	})

	t.Run("update on one object on name and labels fields", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		reg, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		require.NoError(t, ad.Update(ctx, &containerregistry.UpdateRegistryRequest{
			RegistryId: reg.Id,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "labels"}},
			Name:       "reg1_updated",
			Labels:     map[string]string{},
		}))
		res, err := ad.Read(ctx, reg.Id)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, reg1.FolderId, res.FolderId)
		assert.Equal(t, "reg1_updated", res.Name)
		assert.Equal(t, map[string]string{}, res.Labels)
	})

	t.Run("update on one object on no fields", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		reg, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		require.NoError(t, ad.Update(ctx, &containerregistry.UpdateRegistryRequest{
			RegistryId: reg.Id,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{}},
			Name:       "reg1_updated",
			Labels:     map[string]string{},
		}))
		res, err := ad.Read(ctx, reg.Id)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, reg1.FolderId, res.FolderId)
		assert.Equal(t, reg1.Name, res.Name)
		assert.Equal(t, reg1.Labels, res.Labels)
	})

	t.Run("update on no object", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		_, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		err = ad.Update(ctx, &containerregistry.UpdateRegistryRequest{
			RegistryId: "reg-non-existent",
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			Name:       "reg1_updated",
			Labels:     map[string]string{},
		})
		// Assert
		assert.Error(t, err)
	})
}

func TestDelete(t *testing.T) {
	t.Run("delete on one object", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		res, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		require.NoError(t, ad.Delete(ctx, res.Id))
		lst, err := ad.List(ctx, res.FolderId)
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 0)
	})

	t.Run("delete on multiple objects", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		reg2 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg2",
			Labels:   map[string]string{"key": "label"},
		}
		reg3 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg3",
			Labels:   map[string]string{"key": "label"},
		}
		_, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)
		res, err := ad.Create(ctx, &reg2)
		require.NoError(t, err)
		_, err = ad.Create(ctx, &reg3)
		require.NoError(t, err)

		// Act
		require.NoError(t, ad.Delete(ctx, res.Id))
		lst, err := ad.List(ctx, res.FolderId)
		require.NoError(t, err)

		// Assert
		assert.Len(t, lst, 2)
	})

	t.Run("delete on no object", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		ad := NewFakeYandexContainerRegistryAdapter()
		reg1 := containerregistry.CreateRegistryRequest{
			FolderId: "folder",
			Name:     "reg1",
			Labels:   map[string]string{"key": "label"},
		}
		res, err := ad.Create(ctx, &reg1)
		require.NoError(t, err)

		// Act
		err = ad.Delete(ctx, "reg-non-existent-id")
		lst, err2 := ad.List(ctx, res.FolderId)
		require.NoError(t, err2)

		// Assert
		assert.Error(t, err)
		assert.Len(t, lst, 1)
	})
}
