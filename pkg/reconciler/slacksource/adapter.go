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
	"strconv"

	corev1 "k8s.io/api/core/v1"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common"
	"github.com/triggermesh/pkg/reconciler"
	"github.com/triggermesh/pkg/reconciler/resource"
)

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
func adapterServiceBuilder(src *v1alpha1.SlackSource, cfg *adapterConfig) reconciler.AdapterServiceBuilderFunc {
	adapterName := reconciler.AdapterName(src)

	return func(sinkURI *apis.URL) *servingv1.Service {
		name := kmeta.ChildName(adapterName+"-", src.Name)

		var sinkURIStr string
		if sinkURI != nil {
			sinkURIStr = sinkURI.String()
		}

		return resource.NewKnService(src.Namespace, name,
			resource.Controller(src),

			resource.Label(reconciler.AppNameLabel, adapterName),
			resource.Label(reconciler.AppInstanceLabel, src.Name),
			resource.Label(reconciler.AppComponentLabel, reconciler.ComponentAdapter),
			resource.Label(reconciler.AppPartOfLabel, common.PartOf),
			resource.Label(reconciler.AppManagedByLabel, common.ManagedBy),

			resource.PodLabel(reconciler.AppNameLabel, adapterName),
			resource.PodLabel(reconciler.AppInstanceLabel, src.Name),
			resource.PodLabel(reconciler.AppComponentLabel, reconciler.ComponentAdapter),
			resource.PodLabel(reconciler.AppPartOfLabel, common.PartOf),
			resource.PodLabel(reconciler.AppManagedByLabel, common.ManagedBy),

			resource.Image(cfg.Image),

			resource.EnvVar(reconciler.EnvName, src.Name),
			resource.EnvVar(reconciler.EnvNamespace, src.Namespace),
			resource.EnvVar(reconciler.EnvSink, sinkURIStr),
			resource.EnvVars(makeSlackEnvs(src)...),
			resource.EnvVar(reconciler.EnvMetricsPrometheusPort, strconv.Itoa(int(metricsPrometheusPort))),
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
