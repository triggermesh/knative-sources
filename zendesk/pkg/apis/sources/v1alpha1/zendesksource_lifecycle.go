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
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	// ConditionReady has status True when the ZendeskSource is ready to send events.
	ConditionReady = apis.ConditionReady
	// ConditionSinkProvided has status True when the ZendeskSource has been configured with a sink target.
	ConditionSinkProvided apis.ConditionType = "SinkProvided"
	// ConditionDeployed has status True when the ZendeskSource has had it's deployment created.
	ConditionDeployed apis.ConditionType = "Deployed"
	// ConditionTargetCreated has status True when the ZendeskSource has created a Zendesk Target
	ConditionTargetCreated apis.ConditionType = "TargetCreated"
)

// Reasons for status conditions
const (
	// ReasonUnavailable is set on a Deployed condition when an adapter in unavailable.
	ReasonUnavailable = "AdapterUnavailable"

	// ReasonSinkNotFound is set on a SinkProvided condition when a sink does not exist.
	ReasonSinkNotFound = "SinkNotFound"
	// ReasonSinkEmpty is set on a SinkProvided condition when a sink URI is empty.
	ReasonSinkEmpty = "EmptySinkURI"

	// ReasonNoTarget is set on TargetCreated condtion when a Zendesk Target creation failed
	ReasonNoTarget = "ZendeskTargetNotCreated"
)

const (
	// ZendeskSourceEventType is the ZendeskSource CloudEvent type.
	ZendeskSourceEventType = "com.zendesk.events"
)

// ZendeskCondSet is the list of all possible conditions toher than Ready
var ZendeskCondSet = apis.NewLivingConditionSet(
	ConditionSinkProvided,
	ConditionDeployed,
	ConditionTargetCreated,
)

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *ZendeskSourceStatus) InitializeConditions() {
	ZendeskCondSet.Manage(s).InitializeConditions()
}

// PropagateAvailability uses the availability of the adapter to determine whether
// the deployed condition should be marked as true or false.
func (s *ZendeskSourceStatus) PropagateAvailability(ksvc *servingv1.Service) {
	switch {
	case ksvc == nil:
		ZendeskCondSet.Manage(s).MarkUnknown(ConditionDeployed, ReasonUnavailable, "The status Knative Service can not be determined")
		if s.Address != nil {
			s.Address = nil
		}
		return

	case ksvc.Status.IsReady():
		ZendeskCondSet.Manage(s).MarkTrue(ConditionDeployed)

	default:
		ZendeskCondSet.Manage(s).MarkFalse(ConditionDeployed, ReasonUnavailable, "The Knative service %q is unavailable.", ksvc.Name)
	}

	if s.Address == nil {
		s.Address = &duckv1.Addressable{}
	}
	s.Address.URL = ksvc.Status.URL
}

// MarkSink sets the condition that the source has a sink configured.
func (s *ZendeskSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		ZendeskCondSet.Manage(s).MarkTrue(ConditionSinkProvided)
	} else {
		ZendeskCondSet.Manage(s).MarkUnknown(ConditionSinkProvided, ReasonSinkEmpty, "Sink has resolved to empty.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *ZendeskSourceStatus) MarkNoSink(messageFormat string, messageA ...interface{}) {
	ZendeskCondSet.Manage(s).MarkFalse(ConditionSinkProvided, ReasonSinkNotFound, messageFormat, messageA...)
}

// IsReady returns true if the resource is ready overall.
func (s *ZendeskSourceStatus) IsReady() bool {
	return ZendeskCondSet.Manage(s).IsHappy()
}

// MarkNoTargetCreated sets the condition that the source was not able to properly configure a Zendesk Target
func (s *ZendeskSourceStatus) MarkNoTargetCreated(messageFormat string, messageA ...interface{}) {
	ZendeskCondSet.Manage(s).MarkFalse(ConditionTargetCreated, ReasonNoTarget, messageFormat, messageA...)
}

// MarkTargetCreated sets the condition that the source was able to properly configure a Zendesk Target
func (s *ZendeskSourceStatus) MarkTargetCreated() {
	ZendeskCondSet.Manage(s).MarkTrue(ConditionTargetCreated)
}
