package config

import "k8s-connectors/pkg/config"

const (
	FinalizerName = "finalizer.{{ .shortName }}.{{ .groupName }}"
	LongName      = "{{ .longName }}"
	ShortName     = "{{ .shortName }}"
)
