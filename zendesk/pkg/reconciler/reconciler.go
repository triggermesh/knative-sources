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

	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/controller"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	"github.com/nukosuke/go-zendesk/zendesk"
	srcreconciler "github.com/triggermesh/knative-sources/pkg/reconciler"
	"github.com/triggermesh/knative-sources/zendesk/pkg/apis/sources/v1alpha1"
	reconcilerzendesksource "github.com/triggermesh/knative-sources/zendesk/pkg/client/generated/injection/reconciler/sources/v1alpha1/zendesksource"
)

const Title = "TriggermeshExtension"

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

	err = r.createTarget(ctx, src)
	if err != nil {
		src.Status.MarkNoTargetCreated("Could not create a new Zendesk Target: %s", err.Error())
		return controller.NewPermanentError(err)
	}
	src.Status.MarkTargetCreated()

	return event
}

//TODO:
//Replace hardcoding
//Fix MarkNoTarget
// createTarget creates a new zendesk target
func (r *reconciler) createTarget(ctx context.Context, src *v1alpha1.ZendeskSource) error {
	// Does a target exist with the Title? if so return
	check, err := r.checkTargetExistance(Title)
	if err != nil {
		return controller.NewPermanentError(err)
	}
	// if it does not exist : create it
	if !check {
		client, err := zendesk.NewClient(nil)
		if err != nil {
			return err
		}
		if err := client.SetSubdomain("tmdev1"); err != nil {
			return err
		}
		client.SetCredential(zendesk.NewAPITokenCredential("jeff@triggermesh.com", "YU0qskXOY2JT0x0XvxD9II9nfscusjtBNBAf4OFF"))

		t := zendesk.Target{}

		t.TargetURL = "https://zendesksource-triggermesh-zendesksource.jnlasersolutions.dev.munu.io/"
		t.Type = "http_target"
		t.Method = "post"
		t.ContentType = "application/json"
		t.Password = "pass"
		t.Username = "user"
		t.Title = Title

		_, error := client.CreateTarget(ctx, t)
		if error != nil {
			return error
		}

		return nil
	}
	if check {
		return nil
	}
	return nil
}

// checkTargetExistance checks to see if a target with a matching "Title" exisits.
func (r *reconciler) checkTargetExistance(search string) (bool, error) {
	ctx := context.Background()
	client, err := zendesk.NewClient(nil)
	if err != nil {
		return false, err
	}
	if err := client.SetSubdomain("tmdev1"); err != nil {
		return false, err
	}
	client.SetCredential(zendesk.NewAPITokenCredential("jeff@triggermesh.com", "YU0qskXOY2JT0x0XvxD9II9nfscusjtBNBAf4OFF"))
	Target, _, err := client.GetTargets(ctx)
	for _, t := range Target {
		if t.Title == search {
			if t.Active {
				return true, nil
			}
		}
	}
	if err != nil {
		return false, err
	}
	return false, nil
}
