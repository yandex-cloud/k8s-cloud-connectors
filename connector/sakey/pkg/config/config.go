// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package config

import "github.com/yandex-cloud/k8s-cloud-connectors/pkg/config"

const (
	FinalizerName = "finalizer.sakey.connectors.cloud.yandex.com"
	LongName      = "StaticAccessKey"
	ShortName     = "sakey"

	ErrCodeSAKeyNotFound = "yc.sakey.not-found"
)

func GetStaticAccessKeyDescription(clusterName, name string) string {
	return config.CloudClusterLabel + ":" + clusterName + "\n" + config.CloudNameLabel + ":" + name
}
