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

	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/controller"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	"github.com/nukosuke/go-zendesk/zendesk"
	srcreconciler "github.com/triggermesh/knative-sources/pkg/reconciler"
	"github.com/triggermesh/knative-sources/zendesk/pkg/apis/sources/v1alpha1"
	reconcilerzendesksource "github.com/triggermesh/knative-sources/zendesk/pkg/client/generated/injection/reconciler/sources/v1alpha1/zendesksource"
)

const tmTitle = "TriggermeshxExtension"

// Reconciler reconciles a ZendeskSource object
type reconciler struct {
	ksvcr        srcreconciler.KServiceReconciler
	sinkResolver *resolver.URIResolver

	adapterCfg *adapterConfig
}

// reconciler implements Interface
var _ reconcilerzendesksource.Interface = (*reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.ZendeskSource) pkgreconciler.Event {
	src.Status.InitializeConditions()
	src.Status.ObservedGeneration = src.Generation
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

	adapter, event := r.ksvcr.ReconcileKService(ctx, src, makeAdapter(src, r.adapterCfg))
	src.Status.PropagateAvailability(adapter)

	// Prevent the reconciler from trying to create Zendesk integration before spec is avalible
	if src.Spec.Subdomain == "" {
		return event
	}

	// TODO : Add a flag here to skip this
	err = ensureIntegration(ctx, src)
	if err != nil {
		src.Status.MarkNoTargetCreated("Could not create a new Zendesk Target: %s", err.Error())
		return controller.NewPermanentError(err)
	}
	src.Status.MarkTargetCreated()
	return event
}

// ensureIntegration handles all the parts required to create a new webhook integration
func ensureIntegration(ctx context.Context, src *v1alpha1.ZendeskSource) error {

	client, err := zendesk.NewClient(nil)
	if err != nil {
		return err
	}
	if err := client.SetSubdomain(src.Spec.Subdomain); err != nil {
		return err
	}

	client.SetCredential(zendesk.NewAPITokenCredential(src.Spec.Email, src.Spec.Token.SecretKeyRef.Key))

	// Does a target exist with the tm Title? if so return // Todo: Verification of proper URL address
	exists, err := checkTargetExists(ctx, client)
	if err != nil {
		return controller.NewPermanentError(err)
	}

	if !exists {

		t := zendesk.Target{}
		t.TargetURL = "https://zendesksource-zendesksource.jnlasersolutions.dev.munu.io"
		t.Type = "http_target"
		t.Method = "post"
		t.ContentType = "application/json"
		t.Password = src.Spec.Password.SecretKeyRef.Key
		t.Username = src.Spec.Username
		t.Title = tmTitle

		createdTarget, err := client.CreateTarget(ctx, t)
		if err != nil {
			return err
		}

		err = createTrigger(ctx, client, createdTarget)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// checkTargetExists checks if a Zendesk 'Target' with a matching "Title" exists and if the target is active.
// More info on Zendesk Target's: https://developer.zendesk.com/rest_api/docs/support/targets
func checkTargetExists(ctx context.Context, client *zendesk.Client) (bool, error) {

	Target, _, err := client.GetTargets(ctx)
	if err != nil {
		return false, err
	}
	for _, t := range Target {
		if t.Active && t.Title == tmTitle {
			return true, nil
		}
	}
	return false, nil
}

// createTrigger creates a new Zendesk 'Trigger'
// more info on Zendesk 'Trigger's' -> https://developer.zendesk.com/rest_api/docs/support/triggers
func createTrigger(ctx context.Context, client *zendesk.Client, t zendesk.Target) error {

	targetID := strconv.Itoa(int(t.ID))

	ta := zendesk.TriggerAction{
		Field: "notification_target",
		Value: targetID + `,"{"id":"{{ticket.id}}","description":"{{ticket.description}}"}"`,
	}

	var newTrigger = zendesk.Trigger{}
	newTrigger.Title = tmTitle
	newTrigger.Conditions.All = []zendesk.TriggerCondition{{
		Field:    "update_type",
		Operator: "is",
		Value:    "Create",
	}}

	newTrigger.Actions = append(newTrigger.Actions, ta)

	// is there a pre existing trigger that matches the Trigger Actions we need?? if so return
	// more info in Zendesk Trigger Actions -> https://developer.zendesk.com/rest_api/docs/support/triggers#actions
	chk, err := ensureTrigger(ctx, client, newTrigger)
	if err != nil {
		return err
	}

	if chk {
		return nil
	}

	nT, err := client.CreateTrigger(ctx, newTrigger)
	if err != nil {
		return err
	}
	fmt.Println("created trigger:")
	fmt.Println(nT.ID)

	return nil

}

// ensureTrigger see if a Zendesk 'Trigger' with a matching 'Title' exisits & if the 'Trigger' is has the proper URL associated . <-- that part is not done
// more info on Zendesk 'Trigger's' -> https://developer.zendesk.com/rest_api/docs/support/triggers
func ensureTrigger(ctx context.Context, client *zendesk.Client, t zendesk.Trigger) (bool, error) {

	trigggers, _, err := client.GetTriggers(ctx, &zendesk.TriggerListOptions{Active: true})
	if err != nil {
		return false, err
	}

	for _, Trigger := range trigggers {

		if Trigger.Title == tmTitle || Trigger.Actions[0] == t.Actions[0] {
			fmt.Println("Found a matching trigger!")
			fmt.Println(Trigger)
			fmt.Println(Trigger.Title)
			return true, nil

		}

	}

	return false, nil
}
