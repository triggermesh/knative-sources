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

package zendesksource

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgapis "knative.dev/pkg/apis"
	"knative.dev/pkg/controller"

	"github.com/nukosuke/go-zendesk/zendesk"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common/skip"
)

func (r *Reconciler) ensureZendeskTargetAndTrigger(ctx context.Context) error {
	if skip.Skip(ctx) {
		return nil
	}

	src := v1alpha1.SourceFromContext(ctx)
	status := &src.(*v1alpha1.ZendeskSource).Status

	addr := src.GetSourceStatus().Address

	// skip this cycle if the adapter URL wasn't yet determined
	if addr == nil || addr.URL == nil {
		status.MarkTargetNotSynced(v1alpha1.ZendeskReasonNoURL,
			"The receive adapter did not report its public URL yet")
		return nil
	}

	spec := src.(pkgapis.HasSpec).GetUntypedSpec().(v1alpha1.ZendeskSourceSpec)

	apiToken, err := r.secretFrom(ctx, src.GetNamespace(), spec.Token.SecretKeyRef)
	if err != nil {
		status.MarkTargetNotSynced(v1alpha1.ZendeskReasonNoSecret, "Cannot obtain Zendesk API token")
		return err
	}

	webhookPassword, err := r.secretFrom(ctx, src.GetNamespace(), spec.WebhookPassword.SecretKeyRef)
	if err != nil {
		status.MarkTargetNotSynced(v1alpha1.ZendeskReasonNoSecret, "Cannot obtain webhook password")
		return err
	}

	cred := zendesk.NewAPITokenCredential(spec.Email, apiToken)
	client, err := zendesk.NewClient(nil)
	if err != nil {
		return fmt.Errorf("creating Zendesk client: %w", err)
	}
	if err := client.SetSubdomain(spec.Subdomain); err != nil {
		return fmt.Errorf("setting Zendesk subdomain: %w", err)
	}
	client.SetCredential(cred)

	title := "io.triggermesh.zendesksource." + src.GetNamespace() + "." + src.GetName()

	targets, _, err := client.GetTargets(ctx)
	switch {
	case isDenied(err):
		return controller.NewPermanentError(err)

	case err != nil:
		status.MarkTargetNotSynced(v1alpha1.ZendeskReasonFailedSync, "Unable to list Targets")
		return fmt.Errorf("error retrieving Zendesk Targets: %w", err)
	}

	var currentTarget *zendesk.Target
	for _, t := range targets {
		if t.Title == title {
			currentTarget = &t
			break
		}
	}

	// we only require the Target to exist, users can modify its contents
	// after the integration is setup for the first time.
	if currentTarget == nil {
		desiredTarget := zendesk.Target{
			Title:       title,
			Type:        "http_target",
			TargetURL:   addr.URL.String(),
			Method:      "post",
			Username:    spec.WebhookUsername,
			Password:    webhookPassword,
			ContentType: "application/json",
		}

		resp, err := client.CreateTarget(ctx, desiredTarget)
		if err != nil {
			// TODO: It could happen that the target already exists
			// but is in a different page. We will need to support
			// pagination in a future release of this source.
			status.MarkTargetNotSynced(v1alpha1.ZendeskReasonFailedSync, "Unable to create Target")
			return fmt.Errorf("error creating Zendesk target: %w", err)
		}
		currentTarget = &resp
	}

	triggers, _, err := client.GetTriggers(ctx, &zendesk.TriggerListOptions{})
	if err != nil {
		status.MarkTargetNotSynced(v1alpha1.ZendeskReasonFailedSync, "Unable to list Triggers")
		return fmt.Errorf("error retrieving Zendesk Triggers: %w", err)
	}

	var currentTrigger *zendesk.Trigger
	for _, t := range triggers {
		if t.Title == title {
			currentTrigger = &t
			break
		}
	}

	// we only require the Trigger to exist, users can modify its contents
	// after the integration is setup for the first time.
	if currentTrigger == nil {
		desiredTrigger := zendesk.Trigger{
			Title: title,
			Actions: []zendesk.TriggerAction{{
				Field: "notification_target",
				Value: []string{
					strconv.FormatInt(currentTarget.ID, 10),
					triggerPayloadJSON,
				},
			}},
		}
		desiredTrigger.Conditions.All = []zendesk.TriggerCondition{{
			Field:    "update_type",
			Operator: "is",
			Value:    "Create",
		}}

		if _, err := client.CreateTrigger(ctx, desiredTrigger); err != nil {
			status.MarkTargetNotSynced(v1alpha1.ZendeskReasonFailedSync, "Unable to create Trigger")
			return fmt.Errorf("error creating Zendesk target trigger: %w", err)
		}
	}

	status.MarkTargetSynced()

	return nil
}

// secretFrom handles the retrieval of secretes from the within the defined namepace
func (r *Reconciler) secretFrom(ctx context.Context, namespace string, secretKeySelector *corev1.SecretKeySelector) (string, error) {
	secret, err := r.kubeClient.CoreV1().Secrets(namespace).Get(secretKeySelector.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	secretVal, ok := secret.Data[secretKeySelector.Key]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret %q", secretKeySelector.Key, secretKeySelector.Name)
	}
	return string(secretVal), nil
}

// isDenied returns whether the given error indicates that a request was denied
// due to authentication issues.
func isDenied(err error) bool {
	if err, ok := err.(zendesk.Error); ok {
		s := err.Status()
		return s == http.StatusUnauthorized || s == http.StatusForbidden
	}
	return false
}

const triggerPayloadJSON = `{
  "ticket": {
    "id": {{ticket.id}},
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
