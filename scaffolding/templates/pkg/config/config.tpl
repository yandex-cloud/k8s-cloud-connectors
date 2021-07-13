package config

import "k8s-connectors/pkg/config"

const (
	FinalizerName = "finalizer.{{ if .shortName }}{{ .shortName }}{{ else }}{{ "" | regexReplaceAll "[^A-Z]" .longName | lower }}{{ end }}.{{ .groupName }}"
	LongName      = "{{ .longName }}"
	ShortName     = "{{ if .shortName }}{{ .shortName }}{{ else }}{{ "" | regexReplaceAll "[^A-Z]" .longName | lower }}{{ end }}"
)
