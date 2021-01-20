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

package common

import (
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common/resource"
)

const metricsPrometheusPort uint16 = 9092

// AdapterName returns the adapter's name for the given source object.
func AdapterName(o kmeta.OwnerRefable) string {
	return strings.ToLower(o.GetGroupVersionKind().Kind)
}

// NewAdapterDeployment is a wrapper around resource.NewDeployment which
// pre-populates attributes common to all adapter Deployments.
func NewAdapterDeployment(src kmeta.OwnerRefable, sinkURI *apis.URL, opts ...resource.ObjectOption) *appsv1.Deployment {
	app := AdapterName(src)
	meta := src.GetObjectMeta()
	srcNs := meta.GetNamespace()
	srcName := meta.GetName()

	var sinkURIStr string
	if sinkURI != nil {
		sinkURIStr = sinkURI.String()
	}

	return resource.NewDeployment(srcNs, kmeta.ChildName(app+"-", srcName),
		append([]resource.ObjectOption{
			resource.TerminationErrorToLogs,
			resource.Controller(src),

			resource.Label(appNameLabel, app),
			resource.Label(appInstanceLabel, srcName),
			resource.Label(appComponentLabel, componentAdapter),
			resource.Label(appPartOfLabel, partOf),
			resource.Label(appManagedByLabel, managedBy),

			resource.Selector(appNameLabel, app),
			resource.Selector(appInstanceLabel, srcName),
			resource.PodLabel(appComponentLabel, componentAdapter),
			resource.PodLabel(appPartOfLabel, partOf),
			resource.PodLabel(appManagedByLabel, managedBy),

			resource.EnvVar(envNamespace, srcNs),
			resource.EnvVar(envName, srcName),
			resource.EnvVar(envSink, sinkURIStr),
		}, opts...)...,
	)
}

// NewAdapterKnService is a wrapper around resource.NewKnService which
// pre-populates attributes common to all adapter Knative Services.
func NewAdapterKnService(src kmeta.OwnerRefable, sinkURI *apis.URL, opts ...resource.ObjectOption) *servingv1.Service {
	app := AdapterName(src)
	meta := src.GetObjectMeta()
	srcNs := meta.GetNamespace()
	srcName := meta.GetName()

	var sinkURIStr string
	if sinkURI != nil {
		sinkURIStr = sinkURI.String()
	}

	return resource.NewKnService(srcNs, kmeta.ChildName(app+"-", srcName),
		append([]resource.ObjectOption{
			resource.Controller(src),

			resource.Label(appNameLabel, app),
			resource.Label(appInstanceLabel, srcName),
			resource.Label(appComponentLabel, componentAdapter),
			resource.Label(appPartOfLabel, partOf),
			resource.Label(appManagedByLabel, managedBy),

			resource.PodLabel(appNameLabel, app),
			resource.PodLabel(appInstanceLabel, srcName),
			resource.PodLabel(appComponentLabel, componentAdapter),
			resource.PodLabel(appPartOfLabel, partOf),
			resource.PodLabel(appManagedByLabel, managedBy),

			resource.EnvVar(envNamespace, srcNs),
			resource.EnvVar(envName, srcName),
			resource.EnvVar(envSink, sinkURIStr),
			resource.EnvVar(envMetricsPrometheusPort, strconv.FormatUint(uint64(metricsPrometheusPort), 10)),
		}, opts...)...,
	)
}

// MaybeAppendValueFromEnvVar conditionally appends an EnvVar to env based on
// the contents of valueFrom.
// ValueFromSecret takes precedence over Value in case the API didn't reject
// the object despite the CRD's schema validation
func MaybeAppendValueFromEnvVar(envs []corev1.EnvVar, key string, valueFrom v1alpha1.ValueFromField) []corev1.EnvVar {
	if vfs := valueFrom.ValueFromSecret; vfs != nil {
		return append(envs, corev1.EnvVar{
			Name: key,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: vfs,
			},
		})
	}

	if v := valueFrom.Value; v != "" {
		return append(envs, corev1.EnvVar{
			Name:  key,
			Value: v,
		})
	}

	return envs
}
