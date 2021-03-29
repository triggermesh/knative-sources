/*
Copyright (c) 2021 TriggerMesh Inc.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTTPPollerSource is the schema for the event source.
type HTTPPollerSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HTTPPollerSourceSpec `json:"spec,omitempty"`
	Status EventSourceStatus    `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object = (*HTTPPollerSource)(nil)
	_ EventSource    = (*HTTPPollerSource)(nil)
)

// HTTPPollerSourceSpec defines the desired state of the event source.
type HTTPPollerSourceSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// EventType for the event that will be generated.
	EventType string `json:"eventType"`

	// EventSource for the event that will be generated.
	// +optional
	EventSource *string `json:"eventSource,omitempty"`

	// Endpoint to connect to.
	Endpoint apis.URL `json:"endpoint"`

	// Method to use at requests.
	Method string `json:"method"`

	// SkipVerify disables server certificate validation.
	// +optional
	SkipVerify *bool `json:"skipVerify,omitempty"`

	// CACertificate uses the CA certificate to verify the remote server certificate.
	// +optional
	CACertificate *string `json:"caCertificate,omitempty"`

	// BasicAuthUsername used for basic authentication.
	// +optional
	BasicAuthUsername *string `json:"basicAuthUsername,omitempty"`

	// BasicAuthPassword used for basic authentication.
	// +optional
	BasicAuthPassword *ValueFromField `json:"basicAuthPassword,omitempty"`

	// Headers to be included at HTTP requests
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// FrequencySeconds polling the endpoint
	FrequencySeconds int `json:"frequencySeconds,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTTPPollerSourceList contains a list of event sources.
type HTTPPollerSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPPollerSource `json:"items"`
}
