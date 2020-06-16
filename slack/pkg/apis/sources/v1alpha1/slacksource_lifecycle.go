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
	appsv1 "k8s.io/api/apps/v1"
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
)

const (
	// SlackConditionReady has status True when the SlackSource is ready to send events.
	SlackConditionReady = apis.ConditionReady

	// SlackConditionSinkProvided has status True when the SlackSource has been configured with a sink target.
	SlackConditionSinkProvided apis.ConditionType = "SinkProvided"

	// SlackConditionSecretsProvided has status True when the SlackSource has valid secret references
	SlackConditionSecretsProvided apis.ConditionType = "SecretsProvided"

	// SlackConditionDeployed has status True when the SlackSource has had it's deployment created.
	SlackConditionDeployed apis.ConditionType = "Deployed"
)

var SlackCondSet = apis.NewLivingConditionSet(
	SlackConditionSinkProvided,
	SlackConditionSecretsProvided,
	SlackConditionDeployed,
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *SlackSourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return SlackCondSet.Manage(s).GetCondition(t)
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *SlackSourceStatus) InitializeConditions() {
	SlackCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the condition that the source has a sink configured.
func (s *SlackSourceStatus) MarkSink(uri *apis.URL) {
	s.SinkURI = uri
	if len(uri.String()) > 0 {
		SlackCondSet.Manage(s).MarkTrue(SlackConditionSinkProvided)
	} else {
		SlackCondSet.Manage(s).MarkUnknown(SlackConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *SlackSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	SlackCondSet.Manage(s).MarkFalse(SlackConditionSinkProvided, reason, messageFormat, messageA...)
}

// MarkSecrets sets the condition that the source has a valid spec
func (s *SlackSourceStatus) MarkSecrets() {
	SlackCondSet.Manage(s).MarkTrue(SlackConditionSecretsProvided)
}

// MarkNoSecrets sets the condition that the source does not have a valid spec
func (s *SlackSourceStatus) MarkNoSecrets(reason, messageFormat string, messageA ...interface{}) {
	SlackCondSet.Manage(s).MarkFalse(SlackConditionSecretsProvided, reason, messageFormat, messageA...)
}

// PropagateDeploymentAvailability uses the availability of the provided Deployment to determine if
// SlackConditionDeployed should be marked as true or false.
func (s *SlackSourceStatus) PropagateDeploymentAvailability(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		SlackCondSet.Manage(s).MarkTrue(SlackConditionDeployed)
	} else {
		// I don't know how to propagate the status well, so just give the name of the Deployment
		// for now.
		SlackCondSet.Manage(s).MarkFalse(SlackConditionDeployed, "DeploymentUnavailable", "The Deployment '%s' is unavailable.", d.Name)
	}
}

// IsReady returns true if the resource is ready overall.
func (s *SlackSourceStatus) IsReady() bool {
	return SlackCondSet.Manage(s).IsHappy()
}
