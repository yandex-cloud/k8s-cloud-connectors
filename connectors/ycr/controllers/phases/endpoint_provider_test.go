// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import "testing"

func TestEndpointProviderIsUpdated(t *testing.T) {
	t.Run("is updated on configmap existence", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})

	t.Run("is updated on many configmap existence", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})

	t.Run("is not updated on empty cloud", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})

	t.Run("is not updated on other objects existence", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})
}

func TestEndpointProviderUpdate(t *testing.T) {
	t.Run("update on empty cloud creates configmap", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})

	t.Run("update on non-empty cloud creates configmap", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})
}

func TestEndpointProviderCleanup(t *testing.T) {
	t.Run("cleanup on empty cloud does nothing", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})

	t.Run("cleanup on cloud with other configmaps does nothing", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})

	t.Run("cleanup on cloud with this and other configmaps deletes this configmap", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})

	t.Run("cleanup on cloud with this configmap deletes this configmap", func(t *testing.T) {
		// Arrange

		// Act

		// Assert
	})
}
