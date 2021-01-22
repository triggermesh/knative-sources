/*
Copyright (c) 2020-2021 TriggerMesh Inc.

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

import (
	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common/resource"
)

const (
	envZdSubdomain   = "ZENDESK_SUBDOMAIN"
	envZdWebhookUser = "ZENDESK_WEBHOOK_USERNAME"
	envZdWebhookPwd  = "ZENDESK_WEBHOOK_PASSWORD" //nolint:gosec
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/zendesksource-adapter"`

	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// adapterServiceBuilder returns an AdapterServiceBuilderFunc for the
// given source object and adapter config.
func adapterServiceBuilder(src *v1alpha1.ZendeskSource, cfg *adapterConfig) common.AdapterServiceBuilderFunc {
	return func(sinkURI *apis.URL) *servingv1.Service {
		return common.NewAdapterKnService(src, sinkURI,
			resource.Image(cfg.Image),

			resource.EnvVar(envZdSubdomain, src.Spec.Subdomain),
			resource.EnvVar(envZdWebhookUser, src.Spec.WebhookUsername),
			resource.EnvVars(common.MaybeAppendValueFromEnvVar(nil,
				envZdWebhookPwd, src.Spec.WebhookPassword)...,
			),
			resource.EnvVars(cfg.configs.ToEnvVars()...),
		)
	}
}
