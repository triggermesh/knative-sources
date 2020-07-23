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

	"k8s.io/client-go/kubernetes"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/controller"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	"github.com/triggermesh/knative-sources/http/pkg/apis/sources/v1alpha1"
	reconcilerhttpsource "github.com/triggermesh/knative-sources/http/pkg/client/generated/injection/reconciler/sources/v1alpha1/httpsource"
	srcreconciler "github.com/triggermesh/knative-sources/pkg/reconciler"
)

// Reconciler reconciles a HttpSource object
type reconciler struct {
	ksvcr         srcreconciler.KServiceReconciler
	sinkResolver  *resolver.URIResolver
	kubeClientSet kubernetes.Interface

	adapterCfg *adapterConfig
}

// reconciler implements Interface
var _ reconcilerhttpsource.Interface = (*reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.HttpSource) pkgreconciler.Event {
	// TODO add source using the adapter's name
	src.Status.CloudEventAttributes = []duckv1.CloudEventAttributes{{Type: src.Spec.EventType}}

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

	ksvc, event := r.ksvcr.ReconcileKService(ctx, src, makeAdapter(src, r.adapterCfg))
	src.Status.PropagateAvailability(ksvc)

	return event
}
