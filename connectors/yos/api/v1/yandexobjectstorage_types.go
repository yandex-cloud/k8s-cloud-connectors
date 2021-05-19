// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// YandexObjectStorageSpec defines the desired state of YandexObjectStorage
type YandexObjectStorageSpec struct {
	// Name: must be unique in Yandex Cloud. Can consist of lowercase latin letters, dashes, dots and numbers
	// and must be from 3 to 64 characters long. Must be immutable.
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=[a-z0-9][a-z0-9-.]*[a-z0-9]
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// ACL: canned Access Control List to apply to this bucket. Read further
	// about ACL in Yandex Cloud here: https://cloud.yandex.ru/docs/storage/concepts/acl
	// +kubebuilder:validation:Optional
	ACL string `json:"ACL,omitempty"`

	// SAKeyName: specifies name of the Static Access Key that is used to authenticate this
	// Yandex Object Storage in the cloud.
	// +kubebuilder:validation:Required
	SAKeyName string `json:"SAKeyName"`
}

// YandexObjectStorageStatus defines the observed state of YandexObjectStorage
type YandexObjectStorageStatus struct {
	// Bucket can be accessed with just a name and
	// key from secret provided by Static Access Key.
}

// YandexObjectStorage is the Schema for the yandex object storage API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type YandexObjectStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YandexObjectStorageSpec   `json:"spec,omitempty"`
	Status YandexObjectStorageStatus `json:"status,omitempty"`
}

// YandexObjectStorageList contains a list of YandexObjectStorage
// +kubebuilder:object:root=true
type YandexObjectStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []YandexObjectStorage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&YandexObjectStorage{}, &YandexObjectStorageList{})
}
