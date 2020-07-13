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
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/controller"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	"github.com/nukosuke/go-zendesk/zendesk"
	srcreconciler "github.com/triggermesh/knative-sources/pkg/reconciler"
	"github.com/triggermesh/knative-sources/zendesk/pkg/apis/sources/v1alpha1"
	reconcilerzendesksource "github.com/triggermesh/knative-sources/zendesk/pkg/client/generated/injection/reconciler/sources/v1alpha1/zendesksource"
)

// Reconciler reconciles a ZendeskSource object
type reconciler struct {
	ksvcr         srcreconciler.KServiceReconciler
	sinkResolver  *resolver.URIResolver
	kubeClientSet kubernetes.Interface

	adapterCfg *adapterConfig
}

// reconciler implements Interface
var _ reconcilerzendesksource.Interface = (*reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.ZendeskSource) pkgreconciler.Event {

	src.Status.CloudEventAttributes = []duckv1.CloudEventAttributes{{Type: v1alpha1.ZendeskSourceEventType}}

	dest := src.Spec.Sink.DeepCopy()
	if dest.Ref != nil && dest.Ref.Namespace == "" {
		dest.Ref.Namespace = src.Namespace
	}

	uri, err := r.sinkResolver.URIFromDestinationV1(*dest, src)
	if err != nil {
		src.Status.MarkNoSink("Could not resolve sink URI: %v", err)
		return controller.NewPermanentError(err)
	}
	src.Status.MarkSink(uri)

	secretToken, err := r.secretFrom(ctx, src.Namespace, src.Spec.Token.SecretKeyRef)
	if err != nil {
		src.Status.MarkNoToken("Could not find the Zendesk API token secret: %v", err)
		return err
	}

	secretPassword, err := r.secretFrom(ctx, src.Namespace, src.Spec.Password.SecretKeyRef)
	if err != nil {
		src.Status.MarkNoPassword("Could not find the Zendesk password secret: %v", err)
		return err
	}

	src.Status.MarkSecretsFound()

	ksvc, event := r.ksvcr.ReconcileKService(ctx, src, makeAdapter(src, r.adapterCfg))
	src.Status.PropagateAvailability(ksvc)
	if event != nil {
		return event
	}

	if ksvc.Status.URL == nil {
		src.Status.MarkNoZendeskTargetCreated("No URL exposed from service to create the zendesk target integration")
		return nil
	}

	zc, err := createZendeskClient(src.Spec.Subdomain, src.Spec.Email, secretToken)
	if err != nil {
		src.Status.MarkNoZendeskTargetCreated("Cannot create Zendesk client: %v", err)
		return err
	}

	zendeskTarget := zendesk.Target{
		TargetURL:   ksvc.Status.URL.String(),
		Type:        "http_target",
		Method:      "post",
		ContentType: "application/json",
		// TODO replace both with arbitrary values from users
		Username: src.Spec.Username,
		Password: secretPassword,
		Title:    "io.triggermesh." + src.Namespace + "." + src.Name,
	}

	err = ensureZendeskTarget(ctx, zc, zendeskTarget)
	if err != nil {
		src.Status.MarkNoZendeskTargetCreated("Error ensuring Zendesk Target: %v", err)
		return err
	}
	src.Status.MarkZendeskTargetCreated()

	return nil
}

func createZendeskClient(subdomain, email, apiToken string) (*zendesk.Client, error) {
	c, err := zendesk.NewClient(nil)
	if err != nil {
		return nil, err
	}

	if err = c.SetSubdomain(subdomain); err != nil {
		return nil, err
	}

	c.SetCredential(zendesk.NewAPITokenCredential(email, apiToken))
	return c, nil
}

func ensureZendeskTarget(ctx context.Context, client *zendesk.Client, target zendesk.Target) error {
	targets, _, err := client.GetTargets(ctx)
	if err != nil {
		return err
	}

	var t *zendesk.Target
	for i := range targets {
		if targets[i].TargetURL == target.URL {
			t = &targets[i]
			break
		}
	}

	if t == nil {
		existing, err := client.CreateTarget(ctx, target)
		if err != nil {
			// It could happen that the target already exists but is
			// in a different page. We will need to support pagination
			// in a future release of this source.
			return err
		}
		t = &existing
	}

	triggers, _, err := client.GetTriggers(ctx, &zendesk.TriggerListOptions{})
	if err != nil {
		return err
	}

	var tr *zendesk.Trigger
	for i := range triggers {
		if triggers[i].Title == "io.triggermesh.sources" {
			tr = &triggers[i]
			break
		}
	}

	if tr != nil {
		// we only require matching the trigger title, users
		// can modify the trigger contents after the integration
		// is setup for the first time.
		return nil
	}

	// Zendesk trigger. See: https://developer.zendesk.com/rest_api/docs/support/triggers
	trigger := zendesk.Trigger{
		Title:  "io.triggermesh.sources",
		Active: true,
		Actions: []zendesk.TriggerAction{{
			Field: "notification_target",
			Value: []string{
				strconv.FormatInt(t.ID, 10),
				`{"id":"{{ticket.id}}","description":"{{ticket.description}}"}`,
			},
		}},
		Conditions: struct {
			All []zendesk.TriggerCondition `json:"all"`
			Any []zendesk.TriggerCondition `json:"any"`
		}{
			All: []zendesk.TriggerCondition{{
				Field:    "update_type",
				Operator: "is",
				Value:    "Create",
			}},
		},
	}

	if _, err = client.CreateTrigger(ctx, trigger); err != nil {
		return err
	}

	return nil
}

// secretFrom handles the retrieval of secretes from the within the defined namepace
func (r *reconciler) secretFrom(ctx context.Context, namespace string, secretKeySelector *corev1.SecretKeySelector) (string, error) {
	secret, err := r.kubeClientSet.CoreV1().Secrets(namespace).Get(secretKeySelector.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	secretVal, ok := secret.Data[secretKeySelector.Key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret %q", secretKeySelector.Key, secretKeySelector.Name)
	}
	return string(secretVal), nil
}
