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

package slack

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"

	"github.com/triggermesh/knative-sources/slack/pkg/apis/sources/v1alpha1"
	reconcilerslacksource "github.com/triggermesh/knative-sources/slack/pkg/client/generated/injection/reconciler/sources/v1alpha1/slacksource"
	"github.com/triggermesh/knative-sources/slack/pkg/reconciler"
	"github.com/triggermesh/knative-sources/slack/pkg/reconciler/slack/resources"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason SlackSourceReconciled.
func newReconciledNormal(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "SlackSourceReconciled", "SlackSource reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler reconciles a SlackSource object
type Reconciler struct {
	ReceiveAdapterImage string `envconfig:"SLACK_SOURCE_RA_IMAGE" required:"true"`

	kubeClientSet kubernetes.Interface

	dr  *reconciler.DeploymentReconciler
	sbr *reconciler.SinkBindingReconciler
}

// Check that our Reconciler implements Interface
var _ reconcilerslacksource.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.SlackSource) pkgreconciler.Event {
	src.Status.InitializeConditions()
	src.Status.ObservedGeneration = src.Generation

	_, err := r.secretFrom(ctx, src.Namespace, src.Spec.SlackToken.SecretKeyRef)
	if err != nil {
		src.Status.MarkNoSecrets("AccessTokenNotFound", "%s", err)
		return err
	}
	src.Status.MarkSecrets()

	ra, event := r.dr.ReconcileDeployment(ctx, src, resources.MakeReceiveAdapter(&resources.ReceiveAdapterArgs{
		EventSource: src.Namespace + "/" + src.Name,
		Image:       r.ReceiveAdapterImage,
		Source:      src,
		Labels:      resources.Labels(src.Name),
	}))
	if ra != nil {
		src.Status.PropagateDeploymentAvailability(ra)
	}
	if event != nil {
		logging.FromContext(ctx).Infof("returning because event from ReconcileDeployment")
		return event
	}

	if ra != nil {
		logging.FromContext(ctx).Info("going to ReconcileSinkBinding")
		sb, event := r.sbr.ReconcileSinkBinding(ctx, src, src.Spec.SourceSpec, tracker.Reference{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
			Namespace:  ra.Namespace,
			Name:       ra.Name,
		})
		logging.FromContext(ctx).Infof("ReconcileSinkBinding returned %#v", sb)
		if sb != nil {
			src.Status.MarkSink(sb.Status.SinkURI)
		}
		if event != nil {
			return event
		}
	}

	return newReconciledNormal(src.Namespace, src.Name)
}

func (r *Reconciler) secretFrom(ctx context.Context, namespace string, secretKeySelector *corev1.SecretKeySelector) (string, error) {
	secret, err := r.kubeClientSet.CoreV1().Secrets(namespace).Get(secretKeySelector.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	secretVal, ok := secret.Data[secretKeySelector.Key]
	if !ok {
		return "", fmt.Errorf(`key "%s" not found in secret "%s"`, secretKeySelector.Key, secretKeySelector.Name)
	}
	return string(secretVal), nil
}
