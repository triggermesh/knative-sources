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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTTPSource is the schema for the event source.
type HTTPSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HTTPSourceSpec    `json:"spec,omitempty"`
	Status EventSourceStatus `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object = (*HTTPSource)(nil)
	_ EventSource    = (*HTTPSource)(nil)
)

// HTTPSourceSpec defines the desired state of the event source.
type HTTPSourceSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// EventType for the event that will be generated.
	EventType string `json:"eventType"`

	// EventSource for the event that will be generated.
	EventSource *string `json:"eventSource,omitempty"`

	// BasicAuthUsername used for basic authentication.
	// +optional
	BasicAuthUsername *string `json:"basicAuthUsername,omitempty"`

	// BasicAuthPassword used for basic authentication.
	// +optional
	BasicAuthPassword SecretValueFromSource `json:"basicAuthPassword,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HTTPSourceList contains a list of event sources.
type HTTPSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPSource `json:"items"`
}
