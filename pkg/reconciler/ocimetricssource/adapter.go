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

package ocimetricssource

import (
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

const metricsPrometheusPort uint16 = 9092

const (
	oracleApiKey            = "ORACLE_API_PRIVATE_KEY"
	oracleApiKeyPassphrase  = "ORACLE_API_PRIVATE_KEY_PASSPHRASE"
	oracleApiKeyFingerprint = "ORACLE_API_PRIVATE_KEY_FINGERPRINT"
	userOcid                = "ORACLE_USER_OCID"
	tenantOcid              = "ORACLE_TENANT_OCID"
	oracleRegion            = "ORACLE_REGION"
	pollingFrequency        = "ORACLE_METRICS_POLLING_FREQUENCY"
	metricsNamespace        = "ORACLE_METRICS_NAMESPACE"
	metricsQuery            = "ORACLE_METRICS_QUERY"
)

// adapterConfig contains properties used to configure the source's adapter.
// These are automatically populated by envconfig.
type adapterConfig struct {
	// Container image
	Image string `default:"gcr.io/triggermesh/ocimetricssource-adapter"`

	// Configuration accessor for logging/metrics/tracing
	configs source.ConfigAccessor
}

// adapterServiceBuilder returns an AdapterServiceBuilderFunc for the
// given source object and adapter config.
func adapterServiceBuilder(src *v1alpha1.OciMetricsSource, cfg *adapterConfig) common.AdapterServiceBuilderFunc {
	adapterName := common.AdapterName(src)

	return func(sinkURI *apis.URL) *servingv1.Service {
		name := kmeta.ChildName(adapterName+"-", src.Name)

		var sinkURIStr string
		if sinkURI != nil {
			sinkURIStr = sinkURI.String()
		}

		ksvc := resource.NewKnService(src.Namespace, name,
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
			resource.EnvVars(makeOCIMetricsEnvs(src)...),
			resource.EnvVar(common.EnvMetricsPrometheusPort, strconv.Itoa(int(metricsPrometheusPort))),
			resource.EnvVars(cfg.configs.ToEnvVars()...),
		)

		// Due to the polling nature of the OCI Metrics Service, the
		// service must have a minimum of 1 pod running at all times
		templateAnnotations := ksvc.Spec.Template.GetAnnotations()
		if templateAnnotations == nil {
			templateAnnotations = make(map[string]string)
		}

		templateAnnotations["autoscaling.knative.dev/minScale"] = "1"
		ksvc.Spec.Template.SetAnnotations(templateAnnotations)

		return ksvc
	}
}

func makeOCIMetricsEnvs(src *v1alpha1.OciMetricsSource) []corev1.EnvVar {
	ociEnvs := []corev1.EnvVar{{
		Name: oracleApiKey,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: src.Spec.OracleApiPrivateKey.SecretKeyRef,
		},
	},
		{
			Name: oracleApiKeyPassphrase,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.OracleApiPrivateKeyPassphrase.SecretKeyRef,
			},
		},
		{
			Name: oracleApiKeyFingerprint,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.OracleApiPrivateKeyFingerprint.SecretKeyRef,
			},
		},
		{
			Name:  userOcid,
			Value: src.Spec.User,
		},
		{
			Name:  tenantOcid,
			Value: src.Spec.Tenancy,
		},
		{
			Name:  oracleRegion,
			Value: src.Spec.Region,
		},
		{
			Name:  metricsNamespace,
			Value: src.Spec.MetricsNamespace,
		},
		{
			Name:  metricsQuery,
			Value: src.Spec.MetricsQuery,
		}}

	// polling frequency in minutes (default 1m)
	var interval string
	if src.Spec.PollingFrequency != nil {
		interval = *src.Spec.PollingFrequency
	} else {
		interval = "5m"
	}

	ociEnvs = append(ociEnvs, corev1.EnvVar{
		Name:  pollingFrequency,
		Value: interval,
	})

	return ociEnvs
}
