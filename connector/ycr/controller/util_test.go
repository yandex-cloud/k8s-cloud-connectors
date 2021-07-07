// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package controller

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	connectorsv1 "k8s-connectors/connector/ycr/api/v1"
	"k8s-connectors/connector/ycr/controller/adapter"
	"k8s-connectors/pkg/config"
	k8sfake "k8s-connectors/testing/k8s-fake"
	logrfake "k8s-connectors/testing/logr-fake"
)

func setup(t *testing.T) (
	context.Context,
	logr.Logger,
	client.Client,
	adapter.YandexContainerRegistryAdapter,
	yandexContainerRegistryReconciler,
) {
	t.Helper()
	ad := adapter.NewFakeYandexContainerRegistryAdapter()
	cl := k8sfake.NewFakeClient()
	log := logrfake.NewFakeLogger(t)
	return context.Background(), log, cl, &ad, yandexContainerRegistryReconciler{
		cl,
		&ad,
		log,
		"test-cluster",
	}
}

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
	ctx context.Context,
	t *testing.T,
	ad adapter.YandexContainerRegistryAdapter,
	specName,
	folderID,
	metaName,
	clusterName string,
) *containerregistry.Registry {
	t.Helper()
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
