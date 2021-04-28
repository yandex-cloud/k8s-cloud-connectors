// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phases

import (
	connectorsv1 "k8s-connectors/connectors/ycr/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateObject(specName, folderId, metaName, namespace string) connectorsv1.YandexContainerRegistry {
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
