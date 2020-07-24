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

package httpsource

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

const adapterName = "httpsource"

const (
	envHTTPEventType         = "HTTP_EVENT_TYPE"
	envHTTPEventSource       = "HTTP_EVENT_SOURCE"
	envHTTPBasicAuthUsername = "HTTP_BASICAUTH_USERNAME"
	envHTTPBasicAuthPassword = "HTTP_BASICAUTH_PASSWORD"
)

const metricsPrometheusPort uint16 = 9092

// adapterConfig contains properties used to configure the adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/httpsource-adapter"`

	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// adapterServiceBuilder returns an AdapterServiceBuilderFunc for the
// given source object and adapter config.
func adapterServiceBuilder(src *v1alpha1.HTTPSource, cfg *adapterConfig) common.AdapterServiceBuilderFunc {
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
			resource.EnvVars(makeHTTPEnvs(src)...),
			resource.EnvVar(common.EnvMetricsPrometheusPort, strconv.Itoa(int(metricsPrometheusPort))),
			resource.EnvVars(cfg.configs.ToEnvVars()...),
		)
	}
}

func makeHTTPEnvs(src *v1alpha1.HTTPSource) []corev1.EnvVar {
	envs := []corev1.EnvVar{{
		Name:  envHTTPEventType,
		Value: src.Spec.EventType,
	}, {
		Name:  envHTTPEventSource,
		Value: src.AsEventSource(),
	}}

	if user, passref := src.Spec.BasicAuthUsername, src.Spec.BasicAuthPassword.SecretKeyRef; user != nil && passref != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  envHTTPBasicAuthUsername,
			Value: *user,
		}, corev1.EnvVar{
			Name: envHTTPBasicAuthPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: passref,
			}})
	}

	return envs
}
