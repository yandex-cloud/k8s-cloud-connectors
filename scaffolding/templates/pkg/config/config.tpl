package config

import "github.com/yandex-cloud/k8s-cloud-connectors/pkg/config"

const (
	FinalizerName = "finalizer.{{ .shortName }}.{{ .groupName }}"
	LongName      = "{{ .longName }}"
	ShortName     = "{{ .shortName }}"
)
