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

package adapter

// ZendeskEvent contains the event payload
type ZendeskEvent map[string]interface{}

// Type incoming zendesk event for the event
func (e ZendeskEvent) Type() string {
	s, ok := e["type"]
	if !ok {
		return ""
	}
	return s.(string)
}

// ZendeskEventWrapper contains a common wrapper for all Ticket events.
// more information on Zendesk "Ticket" -> https://developer.zendesk.com/rest_api/docs/support/tickets
type ZendeskEventWrapper struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
