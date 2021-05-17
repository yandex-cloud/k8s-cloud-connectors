// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package config

import "k8s-connectors/pkg/config"

const (
	FinalizerName string = "finalizer.sakey.connectors.cloud.yandex.com"
	LongName      string = "StaticAccessKey"
	ShortName     string = "sakey"
)

func GetStaticAccessKeyDescription(clusterName, name string) string {
	return config.CloudClusterLabel + ":" + clusterName + "\n" + config.CloudNameLabel + ":" + name
}
