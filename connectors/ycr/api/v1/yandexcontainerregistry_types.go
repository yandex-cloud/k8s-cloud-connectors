// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RegistryStatus string

const (
	Creating RegistryStatus = "CREATING"
	Active   RegistryStatus = "ACTIVE"
	Deleting RegistryStatus = "DELETING"
)

// YandexContainerRegistrySpec: defines the desired state of YandexContainerRegistry
type YandexContainerRegistrySpec struct {
	// Name: name of registry
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=3
	//+kubebuilder:validation:MaxLength=63
	Name string `json:"name,omitempty"`

	// FolderId: id of a folder in which registry is located
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:
	FolderId string `json:"folderId,omitempty"`
}

// YandexContainerRegistryStatus: defines the observed state of YandexContainerRegistry
type YandexContainerRegistryStatus struct {
	// Id: id of registry
	Id string `json:"id,omitempty"`

	// Status: status of registry.
	// Valid values are:
	// - CREATING
	// - ACTIVE
	// - DELETING
	Status RegistryStatus `json:"status,omitempty"`

	// CreatedAt: RFC3339-formatted string, representing creation time of resource
	CreatedAt string `json:"createdAt,omitempty"`

	// Labels: registry labels in key:value form. Maximum of 64 labels for resource is allowed
	Labels map[string]string `json:"labels,omitempty"`
}

// YandexContainerRegistry: is the Schema for the yandexcontainerregistries API
//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=yc-registry
type YandexContainerRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YandexContainerRegistrySpec   `json:"spec,omitempty"`
	Status YandexContainerRegistryStatus `json:"status,omitempty"`
}

// YandexContainerRegistryList: contains a list of YandexContainerRegistry
//+kubebuilder:object:root=true
type YandexContainerRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []YandexContainerRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&YandexContainerRegistry{}, &YandexContainerRegistryList{})
}
