package {{ .version }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// {{ .longName }}Spec defines the desired state of {{ .longName }}
type {{ .longName }}Spec struct {
    // SpecField: some field in the spec
	SpecField string `json:"specField"`

	// TODO: fill the spec
}

// {{ .longName }}Status defines the observed state of {{ .longName }}
type {{ .longName }}Status struct {
	// StatusField: some field in the status
	StatusField string `json:"statusField,omitempty"`

	// TODO: fill the status
}

// {{ .longName }} is the Schema for the {{ .longName | lower }} API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path={{ .longName }}s,singular={{ .longName }},shortName={{ .shortName }}
type {{ .longName }} struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   {{ .longName }}Spec   `json:"spec,omitempty"`
	Status {{ .longName }}Status `json:"status,omitempty"`
}

// {{ .longName }}List contains a list of {{ .longName }}
// +kubebuilder:object:root=true
type {{ .longName }}List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []{{ .longName }} `json:"items"`
}

func init() {
	SchemeBuilder.Register(&{{ .longName }}{}, &{{ .longName }}List{})
}
