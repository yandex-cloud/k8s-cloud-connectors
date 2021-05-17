// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package phase

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	connectorsv1 "k8s-connectors/connectors/sakey/api/v1"
)

func createObject(saID, metaName, namespace string) connectorsv1.StaticAccessKey {
	return connectorsv1.StaticAccessKey{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      metaName,
		},
		Spec: connectorsv1.StaticAccessKeySpec{
			ServiceAccountID: saID,
		},
	}
}
