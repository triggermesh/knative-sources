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

package reconciler

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/knative-sources/pkg/reconciler/resources"
	"github.com/triggermesh/knative-sources/slack/pkg/apis/sources/v1alpha1"
)

const (
	adapterName = "slacksource"
	partOf      = "slacksource"
	managedBy   = "slacksource-controller"
)

// adapterConfig contains properties used to configure the adapter.
// Public fields are automatically populated by envconfig.
type adapterConfig struct {
	// Configuration accessor for observability logging/metrics/tracing
	obsConfig source.ConfigAccessor

	// Container image
	Image string `envconfig:"SLACKSOURCE_ADAPTER_IMAGE" required:"true"`
}

// MakeAdapter generates the Receive Adapter KService for Slack sources.
func makeAdapter(source *v1alpha1.SlackSource, cfg *adapterConfig) *servingv1.Service {
	name := kmeta.ChildName(adapterName+"-", source.Name)
	labels := makeAdapterLabels(source.Name)
	envSvc := makeServiceEnv(name, source.Namespace)
	envApp := makeAppEnv(&source.Spec)
	envSink := makeSinkEnv(source.Status.SinkURI)
	envObs := makeObsEnv(cfg.obsConfig)
	envs := append(envSvc, envApp...)
	envs = append(envs, envSink...)
	envs = append(envs, envObs...)

	return resources.MakeKService(source.Namespace, name, cfg.Image,
		resources.KsvcLabels(labels),
		resources.KsvcOwner(source),
		resources.KsvcPodLabels(labels),
		resources.KsvcPodEnvVars(envs),
	)
}

func makeServiceEnv(name, namespace string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "NAMESPACE",
			Value: namespace,
		}, {
			Name:  "NAME",
			Value: name,
		},
	}
}

func makeSinkEnv(url *apis.URL) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	if url != nil {
		env = append(env, corev1.EnvVar{
			Name:  "K_SINK",
			Value: url.String(),
		})
	}

	return env
}

func makeAppEnv(spec *v1alpha1.SlackSourceSpec) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	if spec.AppID != nil {
		env = append(env, corev1.EnvVar{
			Name:  "APP_ID",
			Value: *spec.AppID,
		})
	}

	if spec.Token != nil {
		env = append(env, corev1.EnvVar{
			Name: "SLACK_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: spec.Token.SecretKeyRef,
			},
		})
	}

	return env
}

func makeObsEnv(cfg source.ConfigAccessor) []corev1.EnvVar {
	env := cfg.ToEnvVars()

	// port already used by queue proxy
	for i := range env {
		if env[i].Name == source.EnvMetricsCfg {
			env[i].Value = ""
			break
		}
	}

	return env
}

func makeAdapterLabels(name string) labels.Set {
	lbls := labels.Set{
		resources.AppNameLabel:      adapterName,
		resources.AppInstanceLabel:  name,
		resources.AppComponentLabel: resources.AdapterComponent,
		resources.AppPartOfLabel:    partOf,
		resources.AppManagedByLabel: managedBy,
	}

	return lbls
}
