/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	pkgapis "knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SlackSource is the schema for the Slack source
type SlackSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SlackSourceSpec `json:"spec"`
	// +optional
	Status SlackSourceStatus `json:"status,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *SlackSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("SlackSource")
}

var (
	_ runtime.Object      = (*SlackSource)(nil)
	_ kmeta.OwnerRefable  = (*SlackSource)(nil)
	_ pkgapis.Validatable = (*SlackSource)(nil)
	_ pkgapis.Defaultable = (*SlackSource)(nil)
	_ duckv1.KRShaped     = (*SlackSource)(nil)
)

// SlackSourceSpec holds the desired state of the SlackSource (from the client).
type SlackSourceSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// SigningSecret can be set to the value of Slack request signing secret
	// to authenticate callbacks.
	// See: https://api.slack.com/authentication/verifying-requests-from-slack
	// +optional
	SigningSecret *SecretValueFromSource `json:"signingSecret,omitempty"`

	// AppID identifies the Slack application generating this event.
	// It helps identifying the App sourcing events when multiple Slack
	// applications shared an endpoint. See: https://api.slack.com/events-api
	// +optional
	AppID *string `json:"appID,omitempty"`
}

// SecretValueFromSource represents the source of a secret value
type SecretValueFromSource struct {
	// The Secret key to select from.
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// SlackSourceStatus communicates the observed state of the SlackSource (from the controller).
type SlackSourceStatus struct {
	duckv1.SourceStatus  `json:",inline"`
	duckv1.AddressStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SlackSourceList is a list of SlackSource resources
type SlackSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SlackSource `json:"items"`
}
