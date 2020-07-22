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

package zendesksource

// ZendeskEvent contains the event payload
type ZendeskEvent map[string]interface{}

// ID returns the ID of the Zendesk ticket
func (ze *ZendeskEvent) ID() string {
	v, ok := (*ze)["id"]
	if !ok {
		return ""
	}

	id, ok := v.(string)
	if !ok {
		return ""
	}

	return id
}

// Title returns the title of the Zendesk ticket
func (ze *ZendeskEvent) Title() string {
	v, ok := (*ze)["title"]
	if !ok {
		return ""
	}

	t, ok := v.(string)
	if !ok {
		return ""
	}

	return t
}

// Type returns the type of the Zendesk ticket
func (ze *ZendeskEvent) Type() string {
	v, ok := (*ze)["type"]
	if !ok {
		return ""
	}

	t, ok := v.(string)
	if !ok {
		return ""
	}

	return t
}
