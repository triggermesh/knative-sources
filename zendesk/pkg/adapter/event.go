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

import (
	"time"
)

// ZendeskEvent contains the event payload
type ZendeskEvent map[string]interface{}

// ID returns the ID of a Zendesk ticket
func (ze *ZendeskEvent) ID() string {
	id := (*ze)["id"]
	return id.(string)
}

// CreatedAt returns the creation time of a Zendesk ticket
func (ze *ZendeskEvent) CreatedAt() time.Time {
	timeCreated, exists := (*ze)["created_at"]
	if !exists {
		return time.Now()
	}
	t, _ := time.Parse(time.RFC3339, timeCreated.(string))
	return t
}

// Title returns the title of a Zendesk ticket
func (ze *ZendeskEvent) Title() string {
	s := (*ze)["title"]
	return s.(string)
}
