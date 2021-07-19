// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-cloud/k8s-cloud-connectors/connector/ycr/pkg/util"
)

func TestMatchSpec(t *testing.T) {
	t.Run(
		"update matches cloud object with spec of resource", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))
			res, err := rc.allocateResource(ctx, log, &obj)
			require.NoError(t, err)

			// Act
			obj.Spec.Name = "resource-upd"
			require.NoError(t, rc.matchSpec(ctx, log, &obj, res))

			ycr, err := util.GetRegistry(ctx, "", "folder", "obj", "test-cluster", ad)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, "resource-upd", ycr.Name)
		},
	)
}
