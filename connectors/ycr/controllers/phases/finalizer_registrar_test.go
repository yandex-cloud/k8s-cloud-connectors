// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/stretchr/testify/assert"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestRegistrar(t *testing.T) {
	// Arrange
	c := k8sfake.NewFakeClient()
	ctx := context.Background()
	log := logrfake.NewFakeLogger()
	resource := connectorsv1.YandexContainerRegistry{
		TypeMeta: metav1.TypeMeta{
			Kind:       "YandexContainerRegistry",
			APIVersion: "connectors.cloud.yandex.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "resource",
			Namespace:         "default",
			Generation:        0,
			CreationTimestamp: metav1.Time{},
			DeletionTimestamp: nil,
			Labels:            map[string]string{},
			Annotations:       map[string]string{},
			OwnerReferences:   []metav1.OwnerReference{},
			Finalizers:        []string{},
			ClusterName:       "",
			ManagedFields:     []metav1.ManagedFieldsEntry{},
		},
		Spec:   connectorsv1.YandexContainerRegistrySpec{
			Name:     "test-resource",
			FolderId: "TODO",
		},
		Status: connectorsv1.YandexContainerRegistryStatus{},
	}
	phase := FinalizerRegistrar{
		Client: &c,
	}

	// Act
	err1 := c.Create(ctx, &resource)

	check1, err2 := phase.IsUpdated(ctx, &resource)

	err3 := phase.Update(ctx, log, &resource)

	check2, err4 := phase.IsUpdated(ctx, &resource)

	err5 := phase.Cleanup(ctx, log, &resource)

	check3, err6 := phase.IsUpdated(ctx, &resource)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NoError(t, err4)
	assert.NoError(t, err5)
	assert.NoError(t, err6)
	assert.False(t, check1)
	assert.True(t, check2)
	assert.False(t, check3)
}
