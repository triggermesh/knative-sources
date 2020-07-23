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
	pkgapis "knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// GetUntypedSpec implements apis.HasSpec.
func (s *HTTPSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// GetConditionSet implements duckv1.KRShaped.
func (s *HTTPSource) GetConditionSet() pkgapis.ConditionSet {
	return eventSourceConditionSet
}

// GetStatus implements duckv1.KRShaped.
func (s *HTTPSource) GetStatus() *duckv1.Status {
	return &s.Status.Status
}

// GetSink implements EventSource.
func (s *HTTPSource) GetSink() *duckv1.Destination { return &s.Spec.Sink }

// GetSourceStatus implements EventSource.
func (s *HTTPSource) GetSourceStatus() *EventSourceStatus { return &s.Status }

// GetEventTypes implements EventSource.
func (*HTTPSource) GetEventTypes() []string { return nil }

// AsEventSource implements EventSource.
func (*HTTPSource) AsEventSource() string { return "" }
