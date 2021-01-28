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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common/resource"
)

const metricsPrometheusPort uint16 = 9092

// AdapterName returns the adapter's name for the given source object.
func AdapterName(src kmeta.OwnerRefable) string {
	return strings.ToLower(src.GetGroupVersionKind().Kind)
}

// NewAdapterDeployment is a wrapper around resource.NewDeployment which
// pre-populates attributes common to all adapter Deployments.
func NewAdapterDeployment(src v1alpha1.EventSource, sinkURI *apis.URL, opts ...resource.ObjectOption) *appsv1.Deployment {
	app := AdapterName(src)
	srcNs := src.GetNamespace()
	srcName := src.GetName()

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

			resource.ServiceAccount(AdapterRBACObjectsName(src)),

			resource.EnvVar(envSink, sinkURIStr),
		}, opts...)...,
	)
}

// NewAdapterKnService is a wrapper around resource.NewKnService which
// pre-populates attributes common to all adapter Knative Services.
func NewAdapterKnService(src v1alpha1.EventSource, sinkURI *apis.URL, opts ...resource.ObjectOption) *servingv1.Service {
	app := AdapterName(src)
	srcNs := src.GetNamespace()
	srcName := src.GetName()

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

			resource.ServiceAccount(AdapterRBACObjectsName(src)),

			resource.EnvVar(envSink, sinkURIStr),
			resource.EnvVar(envMetricsPrometheusPort, strconv.FormatUint(uint64(metricsPrometheusPort), 10)),
		}, opts...)...,
	)
}

// newServiceAccount returns a ServiceAccount object with its OwnerReferences
// metadata attribute populated from the given owners.
func newServiceAccount(src v1alpha1.EventSource, owners []kmeta.OwnerRefable) *corev1.ServiceAccount {
	ownerRefs := make([]metav1.OwnerReference, len(owners))
	for i, owner := range owners {
		ownerRefs[i] = *kmeta.NewControllerRef(owner)
		ownerRefs[i].Controller = ptr.Bool(false)
	}

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       src.GetNamespace(),
			Name:            AdapterRBACObjectsName(src),
			OwnerReferences: ownerRefs,
			Labels:          RBACObjectLabels(src),
		},
	}

}

// newRoleBinding returns a RoleBinding object that binds a ServiceAccount
// (namespace-scoped) to a ClusterRole (cluster-scoped).
func newRoleBinding(src v1alpha1.EventSource, owner *corev1.ServiceAccount) *rbacv1.RoleBinding {
	crGVK := rbacv1.SchemeGroupVersion.WithKind("ClusterRole")
	saGVK := corev1.SchemeGroupVersion.WithKind("ServiceAccount")

	ns := src.GetNamespace()
	n := AdapterRBACObjectsName(src)

	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      n,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(owner, saGVK),
			},
			Labels: RBACObjectLabels(src),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: crGVK.Group,
			Kind:     crGVK.Kind,
			Name:     n,
		},
		Subjects: []rbacv1.Subject{{
			APIGroup:  saGVK.Group,
			Kind:      saGVK.Kind,
			Namespace: ns,
			Name:      n,
		}},
	}
}

// AdapterRBACObjectsName returns a unique name to apply to all RBAC objects
// for the given source's adapter.
func AdapterRBACObjectsName(src kmeta.OwnerRefable) string {
	return AdapterName(src) + "-" + componentAdapter
}

// RBACObjectLabels returns a set of labels to be applied to reconciled RBAC objects.
func RBACObjectLabels(src kmeta.OwnerRefable) labels.Set {
	return labels.Set{
		appNameLabel:      AdapterName(src),
		appComponentLabel: componentAdapter,
		appPartOfLabel:    partOf,
		appManagedByLabel: managedBy,
	}
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
