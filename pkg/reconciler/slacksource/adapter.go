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

package slacksource

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common/resource"
)

const adapterName = "slacksource"

const (
	envSlackAppID         = "SLACK_APP_ID"
	envSlackSigningSecret = "SLACK_SIGNING_SECRET"
)

const metricsPrometheusPort uint16 = 9092

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/slacksource-adapter"`

	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// adapterServiceBuilder returns an AdapterServiceBuilderFunc for the
// given source object and adapter config.
func adapterServiceBuilder(src *v1alpha1.SlackSource, cfg *adapterConfig) common.AdapterServiceBuilderFunc {
	return func(sinkURI *apis.URL) *servingv1.Service {
		name := kmeta.ChildName(fmt.Sprintf("%s-", adapterName), src.Name)

		var sinkURIStr string
		if sinkURI != nil {
			sinkURIStr = sinkURI.String()
		}

		return resource.NewKnService(src.Namespace, name,
			resource.Controller(src),

			resource.Label(common.AppNameLabel, adapterName),
			resource.Label(common.AppInstanceLabel, src.Name),
			resource.Label(common.AppComponentLabel, common.AdapterComponent),
			resource.Label(common.AppPartOfLabel, common.PartOf),
			resource.Label(common.AppManagedByLabel, common.ManagedBy),

			resource.PodLabel(common.AppNameLabel, adapterName),
			resource.PodLabel(common.AppInstanceLabel, src.Name),
			resource.PodLabel(common.AppComponentLabel, common.AdapterComponent),
			resource.PodLabel(common.AppPartOfLabel, common.PartOf),
			resource.PodLabel(common.AppManagedByLabel, common.ManagedBy),

			resource.Image(cfg.Image),

			resource.EnvVar(common.EnvName, src.Name),
			resource.EnvVar(common.EnvNamespace, src.Namespace),
			resource.EnvVar(common.EnvSink, sinkURIStr),
			resource.EnvVars(makeSlackEnvs(src)...),
			resource.EnvVar(common.EnvMetricsPrometheusPort, strconv.Itoa(int(metricsPrometheusPort))),
			resource.EnvVars(cfg.configs.ToEnvVars()...),
		)
	}
}

func makeSlackEnvs(src *v1alpha1.SlackSource) []corev1.EnvVar {
	var slackEnvs []corev1.EnvVar

	if appID := src.Spec.AppID; appID != nil {
		slackEnvs = append(slackEnvs, corev1.EnvVar{
			Name:  envSlackAppID,
			Value: *appID,
		})
	}

	if signSecret := src.Spec.SigningSecret; signSecret != nil {
		slackEnvs = append(slackEnvs, corev1.EnvVar{
			Name: envSlackSigningSecret,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: signSecret.SecretKeyRef,
			},
		})
	}

	return slackEnvs
}
