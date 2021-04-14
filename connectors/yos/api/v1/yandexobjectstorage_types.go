// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// YandexObjectStorageSpec: defines the desired state of YandexObjectStorage
type YandexObjectStorageSpec struct {
	// Name: must be unique in Yandex Cloud. Can consist of lowercase latin letters, dashes, dots and numbers
	// and must be from 3 to 64 characters long.
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=[a-z0-9][a-z0-9-.]*[a-z0-9]
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Size: size of the bucket in gigabytes. Can be in range from 1Gb to 1024Tb, or 1048576Gb. Must be an integer.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1048576
	// +kubebuilder:validation:Required
	Size int `json:"size"`

	// ReadAccess: public access allows any internet user to download objects from the bucket.
	// +kubebuilder:validation:Enum=private;public
	// +kubebuilder:validation:default=private
	ReadAccess string `json:"readAccess"`

	// ListingAccess: public access allows any internet user to get listing of elements in bucket.
	// +kubebuilder:validation:Enum=private;public
	// +kubebuilder:validation:default=private
	ListingAccess string `json:"listingAccess"`

	// ConfigAccess: public access allows any internet user to read CORS configurations, hosting of static sites and
	// bucket objects lifecycles.
	// +kubebuilder:validation:Enum=private;public
	// +kubebuilder:validation:default=private
	ConfigAccess string `json:"configAccess"`

	// StorageType: specifies which storage type that is used by default for object downloading.
	// Standard storage is dedicated to storages with frequent object queries, while cold storage is better suited
	// for long-term storage of objects with rare read queries.
	// +kubebuilder:validation:Enum=standard;cold
	// +kubebuilder:validation:default=standard
	StorageType string `json:"storageType"`
}

// YandexObjectStorageStatus: defines the observed state of YandexObjectStorage
type YandexObjectStorageStatus struct {
	// TODO (covariance) match status with GET from SDK
}

// YandexObjectStorage: is the Schema for the yandex object storage API
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
type YandexObjectStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YandexObjectStorageSpec   `json:"spec,omitempty"`
	Status YandexObjectStorageStatus `json:"status,omitempty"`
}

// YandexObjectStorageList: contains a list of YandexObjectStorage
//+kubebuilder:object:root=true
type YandexObjectStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []YandexObjectStorage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&YandexObjectStorage{}, &YandexObjectStorageList{})
}
