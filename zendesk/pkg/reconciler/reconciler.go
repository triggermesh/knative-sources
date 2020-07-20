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
	"net/http"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/controller"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	srcreconciler "github.com/triggermesh/knative-sources/pkg/reconciler"
	"github.com/triggermesh/knative-sources/zendesk/pkg/apis/sources/v1alpha1"
	reconcilerzendesksource "github.com/triggermesh/knative-sources/zendesk/pkg/client/generated/injection/reconciler/sources/v1alpha1/zendesksource"
	"github.com/triggermesh/knative-sources/zendesk/pkg/zendesk"
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

	secretPassword, err := r.secretFrom(ctx, src.Namespace, src.Spec.WebhookPassword.SecretKeyRef)
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

	zc := zendesk.NewClient(src.Spec.Email, secretToken, src.Spec.Subdomain, &http.Client{})

	zendeskTarget := &zendesk.Target{
		TargetURL:   ksvc.Status.URL.String(),
		Type:        "http_target",
		Method:      "post",
		ContentType: "application/json",
		Username:    src.Spec.WebhookUsername,
		Password:    secretPassword,
		Title:       "io.triggermesh.zendesksource." + src.Namespace + "." + src.Name,
	}

	err = ensureZendeskTarget(ctx, zc, zendeskTarget)
	if err != nil {
		src.Status.MarkNoZendeskTargetCreated("Error ensuring Zendesk Target: %v", err)
		return err
	}
	src.Status.MarkZendeskTargetCreated()

	return nil
}

func ensureZendeskTarget(ctx context.Context, client zendesk.Client, target *zendesk.Target) error {
	tarwrap, err := client.ListTargets(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving Zendesk targets: %w", err)
	}

	var t *zendesk.Target
	for i := range tarwrap.Targets {
		if tarwrap.Targets[i].Title == target.Title {
			t = &tarwrap.Targets[i]
			break
		}
	}

	if t == nil {
		existing, err := client.CreateTarget(ctx, target)
		if err != nil {
			// It could happen that the target already exists but is
			// in a different page. We will need to support pagination
			// in a future release of this source.
			return fmt.Errorf("error creating Zendesk target: %w", err)
		}
		t = existing
	}

	triwrap, err := client.ListTriggers(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving Zendesk triggers: %w", err)
	}

	var tr *zendesk.Trigger
	for i := range triwrap.Triggers {
		if triwrap.Triggers[i].Title == target.Title {
			tr = &triwrap.Triggers[i]
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
	trigger := &zendesk.Trigger{
		Title: target.Title,
		Actions: []zendesk.TriggerAction{{
			Field: "notification_target",
			Value: []string{
				strconv.FormatInt(t.ID, 10),
				triggerPayloadJSON,
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

	trigger.Conditions.All = []zendesk.TriggerCondition{{
		Field:    "update_type",
		Operator: "is",
		Value:    "Create",
	}}

	if _, err = client.CreateTrigger(ctx, trigger); err != nil {
		return fmt.Errorf("error creating Zendesk target trigger: %w", err)
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

const triggerPayloadJSON = `{
  "ticket": {
    "id": "{{ticket.id}}",
    "external_id": "{{ticket.external_id}}",
    "title": "{{ticket.title}}",
    "url": "{{ticket.url}}",
    "description": "{{ticket.description}}",
    "via": "{{ticket.via}}",
    "status": "{{ticket.status}}",
    "priority": "{{ticket.priority}}",
    "ticket_type": "{{ticket.ticket_type}}",
    "group_name": "{{ticket.group.name}}",
    "brand_name": "{{ticket.brand.name}}",
    "due_date": "{{ticket.due_date}}",
    "account": "{{ticket.account}}",
    "assignee": {
      "email": "{{ticket.assignee.email}}",
      "name": "{{ticket.assignee.name}}",
      "first_name": "{{ticket.assignee.first_name}}",
      "last_name": "{{ticket.assignee.last_name}}"
    },
    "requester": {
      "name": "{{ticket.requester.name}}",
      "first_name": "{{ticket.requester.first_name}}",
      "last_name": "{{ticket.requester.last_name}}",
      "email": "{{ticket.requester.email}}",
      "language": "{{ticket.requester.language}}",
      "phone": "{{ticket.requester.phone}}",
      "external_id": "{{ticket.requester.external_id}}",
      "field": "{{ticket.requester_field}}",
      "details": "{{ticket.requester.details}}"
    },
    "organization": {
      "name": "{{ticket.organization.name}}",
      "external_id": "{{ticket.organization.external_id}}",
      "details": "{{ticket.organization.details}}",
      "notes": "{{ticket.organization.notes}}"
    },
    "ccs": "{{ticket.ccs}}",
    "cc_names": "{{ticket.cc_names}}",
    "tags": "{{ticket.tags}}",
    "current_holiday_name": "{{ticket.current_holiday_name}}",
    "ticket_field_id": "{{ticket.ticket_field_ID}}",
    "ticket_field_option_title_id": "{{ticket.ticket_field_option_title_ID}}"
  },
  "current_user": {
    "name": "{{current_user.name}}",
    "first_name": "{{current_user.first_name}}",
    "email": "{{current_user.email}}",
    "organization": {
      "name": "{{current_user.organization.name}}",
      "notes": "{{current_user.organization.notes}}",
      "details": "{{current_user.organization.details}}"
    },
    "external_id": "{{current_user.external_id}}",
    "phone": "{{current_user.phone}}",
    "details": "{{current_user.details}}",
    "notes": "{{current_user.notes}}",
    "language": "{{current_user.language}}"
  },
  "satisfaction": {
    "current_rating": "{{satisfaction.current_rating}}",
    "current_comment": "{{satisfaction.current_comment}}"
  }
}`
