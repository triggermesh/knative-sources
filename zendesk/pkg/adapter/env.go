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
	"knative.dev/eventing/pkg/adapter/v2"
)

// EnvAccesor for configuration parameters
func EnvAccesor() adapter.EnvConfigAccessor {
	return &envAccessor{}
}

type envAccessor struct {
	adapter.EnvConfig

	Token     string `envconfig:"ZENDESK_TOKEN"`
	Email     string `envconfig:"EMAIL"`     // donmt think i need this here
	Subdomain string `envconfig:"SUBDOMAIN"` // donmt think i need this here
	Username  string `envconfig:"USERNAME"`
	Password  string `envconfig:"PASSWORD"`
}
