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

// OciMetricsSource is the schema for the event source.
type OciMetricsSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OciMetricsSourceSpec `json:"spec,omitempty"`
	Status EventSourceStatus    `json:"status,omitempty"`
}

// Check the interfaces the event source should be implementing.
var (
	_ runtime.Object = (*OciMetricsSource)(nil)
	_ EventSource    = (*OciMetricsSource)(nil)
)

// OciMetricsSourceSpec defines the desired state of the event source.
type OciMetricsSourceSpec struct {
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`

	// Oracle User API private key
	OracleApiPrivateKey SecretValueFromSource `json:"oracleApiPrivateKey"`

	// Oracle User API private key passphrase
	OracleApiPrivateKeyPassphrase SecretValueFromSource `json:"oracleApiPrivateKeyPassphrase"`

	// Oracle User API cert fingerprint
	OracleApiPrivateKeyFingerprint SecretValueFromSource `json:"oracleApiPrivateKeyFingerprint"`

	// Oracle Tenancy OCID
	Tenancy string `json:"oracleTenancy"`

	// Oracle User OCID associated with the API key
	User string `json:"oracleUser"`

	// Oracle Cloud Region
	Region string `json:"oracleRegion"`

	// OCI Metrics Polling Frequency
	PollingFrequency *string `json:"metricsPollingFrequency,omitempty"`

	// Namespace for the query metric to use
	MetricsNamespace string `json:"metricsNamespace"`

	// OCI Metrics Query See https://docs.cloud.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/MetricData
	MetricsQuery string `json:"metricsQuery"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OciMetricsSourceList contains a list of event sources.
type OciMetricsSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OciMetricsSource `json:"items"`
}
