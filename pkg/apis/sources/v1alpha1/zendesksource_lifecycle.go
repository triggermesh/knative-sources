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
	"k8s.io/apimachinery/pkg/runtime/schema"

	pkgapis "knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// GetGroupVersionKind implements kmeta.OwnerRefable.
func (*ZendeskSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ZendeskSource")
}

// GetUntypedSpec implements apis.HasSpec.
func (s *ZendeskSource) GetUntypedSpec() interface{} {
	return s.Spec
}

// GetConditionSet implements duckv1.KRShaped.
func (*ZendeskSource) GetConditionSet() pkgapis.ConditionSet {
	return eventSourceConditionSet
}

// GetStatus implements duckv1.KRShaped.
func (s *ZendeskSource) GetStatus() *duckv1.Status {
	return &s.Status.Status
}

// GetSink implements EventSource.
func (s *ZendeskSource) GetSink() *duckv1.Destination {
	return &s.Spec.Sink
}

// GetSourceStatus implements EventSource.
func (s *ZendeskSource) GetSourceStatus() *EventSourceStatus {
	return &s.Status
}

// AsEventSource implements EventSource.
func (s *ZendeskSource) AsEventSource() string {
	return ZendeskSourceName(s.Spec.Subdomain, s.Name)
}

// ZendeskSourceName returns a unique reference to the source suitable for use
// as as a CloudEvent source.
func ZendeskSourceName(subdomain, name string) string {
	return subdomain + ".zendesk.com/" + name
}

// Supported event types
const (
	// ZendeskTicketCreatedEventType is generated upon creation of a Ticket.
	ZendeskTicketCreatedEventType = "com.zendesk.ticket.created"
)

// GetEventTypes implements EventSource.
func (*ZendeskSource) GetEventTypes() []string {
	return []string{
		ZendeskTicketCreatedEventType,
	}
}
