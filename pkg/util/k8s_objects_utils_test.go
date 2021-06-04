// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package util

import (
	"testing"
)

func TestProvideConfigmap(t *testing.T) {
	t.Run(
		"update on empty cloud creates configmap", func(t *testing.T) {
			// Arrange

			// Act

			// Assert
		},
	)

	t.Run(
		"update on non-empty cluster creates configmap", func(t *testing.T) {
			// Arrange

			// Act

			// Assert

		},
	)
}

func TestRemoveConfigmap(t *testing.T) {
	t.Run(
		"cleanup on cloud with other configmaps does nothing", func(t *testing.T) {
			// Arrange

			// Act

			// Assert

		},
	)

	t.Run(
		"cleanup on cloud with this and other configmaps deletes this configmap", func(t *testing.T) {
			// Arrange

			// Act

			// Assert

		},
	)

	t.Run(
		"cleanup on cloud with this configmap deletes this configmap", func(t *testing.T) {
			// Arrange

			// Act

			// Assert

		},
	)
}
