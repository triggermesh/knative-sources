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

package zendesk

import "time"

// Copied from https://github.com/nukosuke/go-zendesk, with a little tweak
// on Triggers to avoid an issue where the TriggerCondition Value is informed
// as a boolean.

// TriggerCondition zendesk trigger condition
//
// ref: https://developer.zendesk.com/rest_api/docs/core/triggers#conditions-reference
type TriggerCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// TriggerAction is zendesk trigger action
//
// ref: https://developer.zendesk.com/rest_api/docs/core/triggers#actions
type TriggerAction struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

// Trigger is zendesk trigger JSON payload format
//
// ref: https://developer.zendesk.com/rest_api/docs/core/triggers#json-format
type Trigger struct {
	ID         int64  `json:"id,omitempty"`
	Title      string `json:"title"`
	Active     bool   `json:"active,omitempty"`
	Position   int64  `json:"position,omitempty"`
	Conditions struct {
		All []TriggerCondition `json:"all"`
		Any []TriggerCondition `json:"any"`
	} `json:"conditions"`
	Actions     []TriggerAction `json:"actions"`
	Description string          `json:"description,omitempty"`
	CreatedAt   *time.Time      `json:"created_at,omitempty"`
	UpdatedAt   *time.Time      `json:"updated_at,omitempty"`
}

// Target is struct for target payload
type Target struct {
	URL       string     `json:"url,omitempty"`
	ID        int64      `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Active    bool       `json:"active,omitempty"`
	// email_target
	Email   string `json:"email,omitempty"`
	Subject string `json:"subject,omitempty"`
	// http_target
	TargetURL   string `json:"target_url,omitempty"`
	Method      string `json:"method,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

// Page is base struct for resource pagination
type Page struct {
	PreviousPage *string `json:"previous_page"`
	NextPage     *string `json:"next_page"`
	Count        int64   `json:"count"`
}
