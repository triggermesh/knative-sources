/*
Copyright (c) 2020 TriggerMesh, Inc

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

// Package status contains helpers to observe the status of Kubernetes objects.
package status

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// DeploymentPodsWaitingState collects the Pods owned by the given Deployment
// and returns the state of the first observed Pod in a "waiting" state, or nil
// if no Pod is found in that state.
func DeploymentPodsWaitingState(d *appsv1.Deployment,
	pi coreclientv1.PodInterface) (*corev1.ContainerStateWaiting, error) {

	pods, err := pi.List(metav1.ListOptions{LabelSelector: d.Spec.Selector.String()})
	if err != nil {
		return nil, err
	}

	for _, p := range pods.Items {
		if p.Status.Phase == corev1.PodRunning || p.Status.Phase == corev1.PodSucceeded {
			continue
		}

		for _, ps := range p.Status.ContainerStatuses {
			if ws := ps.State.Waiting; ws != nil {
				return ws, nil
			}
		}
	}

	return nil, nil
}
