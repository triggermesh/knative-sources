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
	"knative.dev/pkg/apis"
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

// integration bundles the required items to automate the webhook integration with Zendesk
type integration struct {
	password string
	username string
	title    string
	url      apis.URL

	client *zendesk.Client
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
		src.Status.MarkNoSink("Could not resolve sink URI: %s", err.Error())
		return controller.NewPermanentError(err)
	}

	src.Status.MarkSink(uri)
	ksvc, event := r.ksvcr.ReconcileKService(ctx, src, makeAdapter(src, r.adapterCfg))
	src.Status.PropagateAvailability(ksvc)

	// Prevent the reconciler from trying to create Zendesk integration before spec is avalible
	if src.Spec.Subdomain == "" {
		return event
	}

	secretToken, err := r.secretFrom(ctx, ksvc.Namespace, src.Spec.Token.SecretKeyRef)
	if err != nil {
		src.Status.MarkNoToken("Could not find a Zendesk API token:%s", err.Error())
		return err
	}

	secretPassword, err := r.secretFrom(ctx, src.Namespace, src.Spec.Password.SecretKeyRef)
	if err != nil {
		src.Status.MarkNoPassword("Could not find a Password:%s", err.Error())
		return err
	}

	src.Status.MarkSecretsFound()

	if ksvc.Status.GetCondition(apis.ConditionReady).IsTrue() && ksvc.Status.URL != nil {

		i := &integration{url: *ksvc.Status.URL, username: src.Spec.Username, password: secretPassword, title: ksvc.GetName() + "-" + src.GetNamespace()}
		i.client, err = zendesk.NewClient(nil)
		if err != nil {
			return err
		}

		if err := i.client.SetSubdomain(src.Spec.Subdomain); err != nil {
			return err
		}

		i.client.SetCredential(zendesk.NewAPITokenCredential(src.Spec.Email, secretToken))

		err = i.ensureIntegration(ctx)
		if err != nil {
			src.Status.MarkNoTargetCreated("Could not create a new Zendesk Target: %s", err.Error())
			return err
		}
		src.Status.MarkTargetCreated()
		return event
	}

	return nil
}

// newTarget returns a populated zendesk.Target{}
func (i *integration) newTarget() zendesk.Target {
	return zendesk.Target{
		TargetURL:   i.url.String(),
		Type:        "http_target",
		Method:      "post",
		ContentType: "application/json",
		Password:    i.password,
		Username:    i.username,
		Title:       i.title,
	}
}

// ensureIntegration ensures our Target and Trigger creation
// More info on Zendesk Target's: https://developer.zendesk.com/rest_api/docs/support/targets
func (i *integration) ensureIntegration(ctx context.Context) error {
	t := i.newTarget()

	Target, _, err := i.client.GetTargets(ctx)
	if err != nil {
		return err
	}
	for _, t := range Target {
		if t.TargetURL == i.url.String() {
			if err := i.createTrigger(ctx, t.ID); err != nil {
				return err
			}
			return nil
		}
	}

	createdTarget, err := i.client.CreateTarget(ctx, t)
	if err != nil {
		return err
	}

	if err := i.createTrigger(ctx, createdTarget.ID); err != nil {
		return err
	}
	return nil
}

// createTrigger creates a new Zendesk 'Trigger'
// more info on Zendesk 'Trigger's' -> https://developer.zendesk.com/rest_api/docs/support/triggers
func (i *integration) createTrigger(ctx context.Context, id int64) error {

	var triggerValue []string
	triggerValue = append(triggerValue, strconv.FormatInt(id, 10))
	triggerValue = append(triggerValue, `{"id":"{{ticket.id}}","description":"{{ticket.description}}"}`)

	ta := zendesk.TriggerAction{
		Field: "notification_target",
		Value: triggerValue,
	}

	var newTrigger = zendesk.Trigger{}
	newTrigger.Title = i.title
	newTrigger.Conditions.All = []zendesk.TriggerCondition{{
		Field:    "update_type",
		Operator: "is",
		Value:    "Create",
	}}

	// more info in Zendesk Trigger Actions -> https://developer.zendesk.com/rest_api/docs/support/triggers#actions
	newTrigger.Actions = append(newTrigger.Actions, ta)
	exists, err := i.ensureTrigger(ctx, newTrigger)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	if _, err = i.client.CreateTrigger(ctx, newTrigger); err != nil {
		return err
	}

	return nil
}

// ensureTrigger verifies and or creates the Zendesk webhook integration.
// more info on Zendesk 'Trigger's' -> https://developer.zendesk.com/rest_api/docs/support/triggers
func (i *integration) ensureTrigger(ctx context.Context, t zendesk.Trigger) (bool, error) {
	trigggers, _, err := i.client.GetTriggers(ctx, &zendesk.TriggerListOptions{Active: true})
	if err != nil {
		return false, err
	}

	for _, Trigger := range trigggers {
		if Trigger.Title == i.title || Trigger.Actions[0] == t.Actions[0] {
			return true, nil
		}
	}
	return false, nil
}

// secretFrom handles the retrieval of secretes from the within the defined namepace
func (r *reconciler) secretFrom(ctx context.Context, namespace string, secretKeySelector *corev1.SecretKeySelector) (string, error) {
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
