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
	// SlackSourceEventType is the SlackSource CloudEvent type.
	SlackSourceEventType = "com.slack.events"
)

// SlackCondSet is the list of all possible conditions toher than Ready
var SlackCondSet = pkgapis.NewLivingConditionSet(
	ConditionSinkProvided,
	ConditionDeployed,
)

// GetConditionSet implements duckv1.KRShaped.
func (s *SlackSource) GetConditionSet() pkgapis.ConditionSet {
	return SlackCondSet
}

// GetStatus implements duckv1.KRShaped.
func (s *SlackSource) GetStatus() *duckv1.Status {
	return &s.Status.Status
}

// GetSink implements EventSource.
func (*SlackSource) GetSink() *duckv1.Destination { return nil }

// GetSourceStatus implements EventSource.
func (*SlackSource) GetSourceStatus() *EventSourceStatus { return nil }

// GetEventTypes implements EventSource.
func (*SlackSource) GetEventTypes() []string { return nil }

// AsEventSource implements EventSource.
func (*SlackSource) AsEventSource() string { return "" }

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *SlackSourceStatus) InitializeConditions() {
	SlackCondSet.Manage(s).InitializeConditions()
}

// PropagateAvailability uses the availability of the adapter to determine whether
// the deployed condition should be marked as true or false.
func (s *SlackSourceStatus) PropagateAvailability(ksvc *servingv1.Service) {
	switch {
	case ksvc == nil:
		SlackCondSet.Manage(s).MarkUnknown(ConditionDeployed, ReasonUnavailable, "The status Knative Service can not be determined")
		if s.Address != nil {
			s.Address = nil
		}
		return

	case ksvc.IsReady():
		SlackCondSet.Manage(s).MarkTrue(ConditionDeployed)

	default:
		SlackCondSet.Manage(s).MarkFalse(ConditionDeployed, ReasonUnavailable, "The Knative service %q is unavailable.", ksvc.Name)
	}

	if s.Address == nil {
		s.Address = &duckv1.Addressable{}
	}
	s.Address.URL = ksvc.Status.URL
}

// MarkSink sets the condition that the source has a sink configured.
func (s *SlackSourceStatus) MarkSink(uri *pkgapis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		SlackCondSet.Manage(s).MarkTrue(ConditionSinkProvided)
	} else {
		SlackCondSet.Manage(s).MarkUnknown(ConditionSinkProvided, ReasonSinkEmpty, "Sink has resolved to empty.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *SlackSourceStatus) MarkNoSink(messageFormat string, messageA ...interface{}) {
	SlackCondSet.Manage(s).MarkFalse(ConditionSinkProvided, ReasonSinkNotFound, messageFormat, messageA...)
}

// IsReady returns true if the resource is ready overall.
func (s *SlackSourceStatus) IsReady() bool {
	return SlackCondSet.Manage(s).IsHappy()
}
