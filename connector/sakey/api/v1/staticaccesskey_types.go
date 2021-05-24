// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StaticAccessKeySpec defines the desired state of StaticAccessKeySpec
type StaticAccessKeySpec struct {
	// ServiceAccountID: id of service account from which
	// the key will be issued
	// +kubebuilder:validation:Required
	ServiceAccountID string `json:"serviceAccountId"`
}

// StaticAccessKeyStatus defines the observed state of StaticAccessKey
type StaticAccessKeyStatus struct {
	// KeyID: id of an issued key
	KeyID string `json:"keyId,omitempty"`

	// SecretRef: reference to a secret containing
	// issued key values. It is always in the same
	// namespace as the StaticAccessKey.
	SecretName string `json:"secretName,omitempty"`
}

// StaticAccessKey is the Schema for the staticaccesskey API
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=sakey
// +kubebuilder:subresource:status
type StaticAccessKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StaticAccessKeySpec   `json:"spec,omitempty"`
	Status StaticAccessKeyStatus `json:"status,omitempty"`
}

// StaticAccessKeyList contains a list of StaticAccessKey
// +kubebuilder:object:root=true
type StaticAccessKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StaticAccessKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StaticAccessKey{}, &StaticAccessKeyList{})
}
