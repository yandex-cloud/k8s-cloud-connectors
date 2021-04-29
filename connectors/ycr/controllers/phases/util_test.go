// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	"k8s-connectors/connectors/ycr/controllers/adapter"
	"k8s-connectors/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func createObject(specName, folderId, metaName, namespace string) connectorsv1.YandexContainerRegistry {
	return connectorsv1.YandexContainerRegistry{
		Spec: connectorsv1.YandexContainerRegistrySpec{
			Name:     specName,
			FolderId: folderId,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      metaName,
			Namespace: namespace,
		},
	}
}

func createResourceRequireNoError(ctx context.Context, ad adapter.YandexContainerRegistryAdapter, t *testing.T, specName, folderId, metaName, clusterName string) *containerregistry.Registry {
	res, err := ad.Create(ctx, &containerregistry.CreateRegistryRequest{
		FolderId: folderId,
		Name:     specName,
		Labels: map[string]string{
			config.CloudClusterLabel: clusterName,
			config.CloudNameLabel:    metaName,
		},
	})
	require.NoError(t, err)
	return res
}
