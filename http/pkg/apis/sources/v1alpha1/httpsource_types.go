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

// HttpSource is the schema for the Http source
type HttpSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec HttpSourceSpec `json:"spec"`
	// +optional
	Status HttpSourceStatus `json:"status,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *HttpSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("HttpSource")
}

var (
	_ runtime.Object      = (*HttpSource)(nil)
	_ kmeta.OwnerRefable  = (*HttpSource)(nil)
	_ pkgapis.Validatable = (*HttpSource)(nil)
	_ pkgapis.Defaultable = (*HttpSource)(nil)
	_ duckv1.KRShaped     = (*HttpSource)(nil)
)

// HttpSourceSpec holds the desired state of the HttpSource (from the client).
type HttpSourceSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// EventType for the event that will be ingested.
	EventType string `json:"eventType,omitempty"`

	// EventSource for the event that will be ingested.
	EventSource *string `json:"eventSource,omitempty"`

	// BasicAuthUsername used for basic authentication.
	// +optional
	BasicAuthUsername *string `json:"basicAuthUsername,omitempty"`

	// BasicAuthPassword used for basic authentication.
	// +optional
	BasicAuthPassword SecretValueFromSource `json:"basicAuthPassword,omitempty"`
}

// SecretValueFromSource represents the source of a secret value
type SecretValueFromSource struct {
	// The Secret key to select from.
	// +optional
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// HttpSourceStatus communicates the observed state of the HttpSource (from the controller).
type HttpSourceStatus struct {
	duckv1.SourceStatus  `json:",inline"`
	duckv1.AddressStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HttpSourceList is a list of HttpSource resources
type HttpSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []HttpSource `json:"items"`
}
