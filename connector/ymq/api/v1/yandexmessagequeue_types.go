// Copyright (c) 2021 Yandex LLC. All rights reserved.
// Author: Martynov Pavel <covariance@yandex-team.ru>

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// YandexMessageQueueSpec defines the desired state of YandexMessageQueue
type YandexMessageQueueSpec struct {
	// Name: must be unique in Yandex Cloud. Can consist of lowercase latin letters, dashes, dots and numbers
	// and must be up to 80 characters long. Name of FIFO queue must end with ".fifo". Must be immutable.
	// +kubebuilder:validation:MaxLength=80
	// +kubebuilder:validation:Pattern=[a-z0-9][a-z0-9-_]*[a-z0-9]
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// FifoQueue: flag that states whether queue is FIFO or not. Must be immutable.
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	FifoQueue bool `json:"fifoQueue"`

	// ContentBasedDeduplication: flag that enables deduplication by message contents.
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	ContentBasedDeduplication bool `json:"contentBasedDeduplication"`

	// DelaySeconds: Time in seconds for which messages are hidden after sending.
	// Can be from 0 to 900 seconds (15 minutes). Defaults to 0.
	// +kubebuilder:default=0
	// +kubebuilder:validation:Optional
	DelaySeconds int `json:"delaySeconds"`

	// MaximumMessageSize: maximal size of message in bytes. Can vary from 1024 (1 KiB) to 262144 bytes (256 KiB).
	// Defaults to 262144 (256 KiB).
	// +kubebuilder:default=262144
	// +kubebuilder:validation:Optional
	MaximumMessageSize int `json:"maximumMessageSize"`

	// MessageRetentionPeriod: duration of message storing. Can vary from 60 seconds (1 minute) to 1209600 seconds
	// (14 days). Defaults to: 345600 (4 days).
	// +kubebuilder:default=345600
	// +kubebuilder:validation:Optional
	MessageRetentionPeriod int `json:"messageRetentionPeriod"`

	// ReceiveMessageWaitTimeSeconds: timeout for method "ReceiveMessage" measured in seconds. Can vary from
	// 0 to 20 seconds. Defaults to 0.
	// +kubebuilder:default=0
	// +kubebuilder:validation:Optional
	ReceiveMessageWaitTimeSeconds int `json:"receiveMessageWaitTimeSeconds"`

	// TODO (covariance) include redrive policy into spec
	// RedrivePolicy: policy of redirecting messages to DeadLetterQueue. Type of this queue and DLQ must both
	// be either FIFO or not FIFO.

	// VisibilityTimeout: timeout of messages visibility timeout. Can vary from 0 to 43000 seconds. Defaults to 30.
	// +kubebuilder:default=30
	// +kubebuilder:validation:Optional
	VisibilityTimeout int `json:"visibilityTimeout"`

	// SAKeyName: specifies name of the Static Access Key that is used to authenticate this
	// Yandex Object Storage in the cloud.
	// +kubebuilder:validation:Required
	SAKeyName string `json:"SAKeyName"`
}

// YandexMessageQueueStatus defines the observed state of YandexMessageQueue
type YandexMessageQueueStatus struct {
	// URL of created queue
	QueueURL string `json:"queueUrl,omitempty"`
}

// YandexMessageQueue is the Schema for the yandex object storage API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type YandexMessageQueue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YandexMessageQueueSpec   `json:"spec,omitempty"`
	Status YandexMessageQueueStatus `json:"status,omitempty"`
}

// YandexMessageQueueList contains a list of YandexMessageQueue
// +kubebuilder:object:root=true
type YandexMessageQueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []YandexMessageQueue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&YandexMessageQueue{}, &YandexMessageQueueList{})
}
