// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	"k8s-connectors/connector/ycr/controller/adapter"
	"k8s-connectors/pkg/config"
)

func createObject(specName, folderID, metaName, namespace string) connectorsv1.YandexContainerRegistry {
	return connectorsv1.YandexContainerRegistry{
		Spec: connectorsv1.YandexContainerRegistrySpec{
			Name:     specName,
			FolderID: folderID,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      metaName,
			Namespace: namespace,
		},
	}
}

func createResourceRequireNoError(
	ctx context.Context, ad adapter.YandexContainerRegistryAdapter, t *testing.T,
	specName, folderID, metaName, clusterName string,
) *containerregistry.Registry {
	res, err := ad.Create(
		ctx, &containerregistry.CreateRegistryRequest{
			FolderId: folderID,
			Name:     specName,
			Labels: map[string]string{
				config.CloudClusterLabel: clusterName,
				config.CloudNameLabel:    metaName,
			},
		},
	)
	require.NoError(t, err)
	return res
}
