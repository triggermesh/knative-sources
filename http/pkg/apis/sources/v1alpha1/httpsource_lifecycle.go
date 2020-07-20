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
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	// ConditionReady has status True when the HttpSource is ready to send events.
	ConditionReady = pkgapis.ConditionReady

	// ConditionSinkProvided has status True when the HttpSource has been configured with a sink target.
	ConditionSinkProvided pkgapis.ConditionType = "SinkProvided"

	// ConditionDeployed has status True when the HttpSource has had it's deployment created.
	ConditionDeployed pkgapis.ConditionType = "Deployed"
)

// Reasons for status conditions
const (
	// ReasonUnavailable is set on a Deployed condition when an adapter in unavailable.
	ReasonUnavailable = "AdapterUnavailable"

	// ReasonSinkNotFound is set on a SinkProvided condition when a sink does not exist.
	ReasonSinkNotFound = "SinkNotFound"

	// ReasonSinkEmpty is set on a SinkProvided condition when a sink URI is empty.
	ReasonSinkEmpty = "EmptySinkURI"
)

// HttpCondSet is the list of all possible conditions other than 'Ready'
var HttpCondSet = pkgapis.NewLivingConditionSet(
	ConditionSinkProvided,
	ConditionDeployed,
)

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *HttpSourceStatus) InitializeConditions() {
	HttpCondSet.Manage(s).InitializeConditions()
}

// PropagateAvailability uses the availability of the adapter to determine whether
// the deployed condition should be marked as 'true' or 'false'.
func (s *HttpSourceStatus) PropagateAvailability(ksvc *servingv1.Service) {
	switch {
	case ksvc == nil:
		HttpCondSet.Manage(s).MarkUnknown(ConditionDeployed, ReasonUnavailable, "The status Knative Service can not be determined")
		if s.Address != nil {
			s.Address = nil
		}
		return

	case ksvc.IsReady():
		HttpCondSet.Manage(s).MarkTrue(ConditionDeployed)

	default:
		HttpCondSet.Manage(s).MarkFalse(ConditionDeployed, ReasonUnavailable, "The Knative service %q is unavailable.", ksvc.Name)
	}

	if s.Address == nil {
		s.Address = &duckv1.Addressable{}
	}
	s.Address.URL = ksvc.Status.URL
}

// MarkSink sets the condition that the source has a sink configured.
func (s *HttpSourceStatus) MarkSink(uri *pkgapis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		HttpCondSet.Manage(s).MarkTrue(ConditionSinkProvided)
	} else {
		HttpCondSet.Manage(s).MarkUnknown(ConditionSinkProvided, ReasonSinkEmpty, "Sink has resolved to empty.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *HttpSourceStatus) MarkNoSink(messageFormat string, messageA ...interface{}) {
	HttpCondSet.Manage(s).MarkFalse(ConditionSinkProvided, ReasonSinkNotFound, messageFormat, messageA...)
}

// IsReady returns true if the resource is ready overall.
func (s *HttpSourceStatus) IsReady() bool {
	return HttpCondSet.Manage(s).IsHappy()
}

// GetConditionSet implements duckv1.KRShaped.
func (s *HttpSource) GetConditionSet() pkgapis.ConditionSet {
	return HttpCondSet
}

// GetStatus implements duckv1.KRShaped.
func (s *HttpSource) GetStatus() *duckv1.Status {
	return &s.Status.Status
}
