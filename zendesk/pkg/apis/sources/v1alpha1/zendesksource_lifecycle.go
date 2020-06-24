/*
Copyright 2020 The Knative Authors.

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
	appsv1 "k8s.io/api/apps/v1"
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
)

const (
	// ZendeskConditionReady has status True when the ZendeskSource is ready to send events.
	ZendeskConditionReady = apis.ConditionReady

	// ZendeskConditionSinkProvided has status True when the ZendeskSource has been configured with a sink target.
	ZendeskConditionSinkProvided apis.ConditionType = "SinkProvided"

	// ZendeskConditionSecretsProvided has status True when the ZendeskSource has valid secret references
	ZendeskConditionSecretsProvided apis.ConditionType = "SecretsProvided"

	// ZendeskConditionDeployed has status True when the ZendeskSource has had it's deployment created.
	ZendeskConditionDeployed apis.ConditionType = "Deployed"
)

var ZendeskCondSet = apis.NewLivingConditionSet(
	ZendeskConditionSinkProvided,
	ZendeskConditionSecretsProvided,
	ZendeskConditionDeployed,
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *ZendeskSourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return ZendeskCondSet.Manage(s).GetCondition(t)
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *ZendeskSourceStatus) InitializeConditions() {
	ZendeskCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the condition that the source has a sink configured.
func (s *ZendeskSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		ZendeskCondSet.Manage(s).MarkTrue(ZendeskConditionSinkProvided)
	} else {
		ZendeskCondSet.Manage(s).MarkUnknown(ZendeskConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *ZendeskSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	ZendeskCondSet.Manage(s).MarkFalse(ZendeskConditionSinkProvided, reason, messageFormat, messageA...)
}

// MarkSecrets sets the condition that the source has a valid spec
func (s *ZendeskSourceStatus) MarkSecrets() {
	ZendeskCondSet.Manage(s).MarkTrue(ZendeskConditionSecretsProvided)
}

// MarkNoSecrets sets the condition that the source does not have a valid spec
func (s *ZendeskSourceStatus) MarkNoSecrets(reason, messageFormat string, messageA ...interface{}) {
	ZendeskCondSet.Manage(s).MarkFalse(ZendeskConditionSecretsProvided, reason, messageFormat, messageA...)
}

// PropagateDeploymentAvailability uses the availability of the provided Deployment to determine if
// ZendeskConditionDeployed should be marked as true or false.
func (s *ZendeskSourceStatus) PropagateDeploymentAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		ZendeskCondSet.Manage(s).MarkTrue(ZendeskConditionDeployed)
	} else {
		// I don't know how to propagate the status well, so just give the name of the Deployment
		// for now.
		ZendeskCondSet.Manage(s).MarkFalse(ZendeskConditionDeployed, "DeploymentUnavailable", "The Deployment '%s' is unavailable.", d.Name)
	}
}

// IsReady returns true if the resource is ready overall.
func (s *ZendeskSourceStatus) IsReady() bool {
	return ZendeskCondSet.Manage(s).IsHappy()
}
