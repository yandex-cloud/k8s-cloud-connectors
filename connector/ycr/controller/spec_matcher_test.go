// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s-connectors/connector/ycr/pkg/util"
)

func TestSpecMatcherUpdate(t *testing.T) {
	t.Run(
		"update matches cloud object with spec of resource", func(t *testing.T) {
			// Arrange
			ctx, log, cl, ad, rc := setup(t)
			obj := createObject("resource", "folder", "obj", "default")
			require.NoError(t, cl.Create(ctx, &obj))
			require.NoError(t, rc.allocateResource(ctx, log, &obj))

			// Act
			obj.Spec.Name = "resource-upd"
			require.NoError(t, rc.matchSpec(ctx, log, &obj))

			ycr, err := util.GetRegistry(ctx, "", "folder", "obj", "test-cluster", ad)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, "resource-upd", ycr.Name)
		},
	)
}
