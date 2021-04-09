// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSAccessKeySpec: defines the desired state of AWSAccessKeySpec
type AWSAccessKeySpec struct {
	// ServiceAccountID: id of service account from which
	// the key will be issued
	//+kubebuilder:validation:Required
	ServiceAccountID string `json:"name,omitempty"`
}

// AWSAccessKeyStatus: defines the observed state of AWSAccessKey
type AWSAccessKeyStatus struct {
	// KeyID: id of an issued key
	KeyID string `json:"folderId,omitempty"`

	// SecretRef: reference to a secret containing
	// issued key values. It is always in the same
	// namespace as the AWSAccessKey.
	SecretName string `json:"secretName,omitempty"`
}

// AWSAccessKey: is the Schema for the awsaccesskey API
//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=awskey
type AWSAccessKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSAccessKeySpec   `json:"spec,omitempty"`
	Status AWSAccessKeyStatus `json:"status,omitempty"`
}

// AWSAccessKeyList: contains a list of AWSAccessKey
//+kubebuilder:object:root=true
type AWSAccessKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSAccessKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSAccessKey{}, &AWSAccessKeyList{})
}
