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

import "time"

// ZendeskEvent contains the event payload
type ZendeskEvent map[string]interface{}

type CustomField struct {
	ID string `json:"id"`
	// Valid types are string or []string.
	Value interface{} `json:"value"`
}

// ZendeskEventWrapper contains a common wrapper for all events.
type ZendeskEventWrapper struct {
	AdditionalProperties map[string]interface{} `json:"-,omitempty"`

	ID              string        `json:"id,omitempty"`
	Title           string        `json:"title,omitempty"`
	Description     string        `json:"description,omitempty"`
	Status          string        `json:"status,omitempty"`
	CreatedAt       time.Time     `json:"created_at,omitempty"`
	URL             string        `json:"url,omitempty"`
	ExternalID      string        `json:"external_id,omitempty"`
	Type            string        `json:"type,omitempty"`
	Subject         string        `json:"subject,omitempty"`
	RawSubject      string        `json:"raw_subject,omitempty"`
	Priority        string        `json:"priority,omitempty"`
	Recipient       string        `json:"recipient,omitempty"`
	RequesterID     string        `json:"requester_id,omitempty"`
	SubmitterID     string        `json:"submitter_id,omitempty"`
	AssigneeID      string        `json:"assignee_id,omitempty"`
	OrganizationID  string        `json:"organization_id,omitempty"`
	GroupID         string        `json:"group_id,omitempty"`
	CollaboratorIDs []string      `json:"collaborator_ids,omitempty"`
	FollowerIDs     []string      `json:"follower_ids,omitempty"`
	EmailCCIDs      []string      `json:"email_cc_ids,omitempty"`
	ForumTopicID    string        `json:"forum_topic_id,omitempty"`
	ProblemID       string        `json:"problem_id,omitempty"`
	DueAt           time.Time     `json:"due_at,omitempty"`
	Tags            []string      `json:"tags,omitempty"`
	CustomFields    []CustomField `json:"custom_fields,omitempty"`

	Via struct {
		Channel string `json:"channel,omitempty"`
		Source  struct {
			From map[string]interface{} `json:"from,omitempty"`
			To   map[string]interface{} `json:"to,omitempty"`
			Rel  string                 `json:"rel,omitempty"`
		} `json:"source"`
	} `json:"via"`

	SatisfactionRating struct {
		ID      string `json:"id,omitempty"`
		Score   string `json:"score,omitempty"`
		Comment string `json:"comment,omitempty"`
	} `json:"satisfaction_rating,omitempty"`
}

// Type for the event
func (e ZendeskEvent) Type() string {
	s, ok := e["type"]
	if !ok {
		return ""
	}
	return s.(string)
}
