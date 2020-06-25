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

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/tools/cache"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"
	kserviceclient "knative.dev/serving/pkg/client/injection/client"
	kserviceinformer "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"

	srcreconciler "github.com/triggermesh/knative-sources/pkg/reconciler"
	"github.com/triggermesh/knative-sources/slack/pkg/apis/sources/v1alpha1"
	slacksourceinformer "github.com/triggermesh/knative-sources/slack/pkg/client/generated/injection/informers/sources/v1alpha1/slacksource"
	"github.com/triggermesh/knative-sources/slack/pkg/client/generated/injection/reconciler/sources/v1alpha1/slacksource"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	adapterCfg := &adapterConfig{
		obsConfig: source.WatchConfigurations(ctx, adapterName, cmw, source.WithLogging, source.WithMetrics),
	}

	if err := envconfig.Process("", adapterCfg); err != nil {
		logging.FromContext(ctx).Panic(err)
	}

	ksvcInformer := kserviceinformer.Get(ctx)
	slackSourceInformer := slacksourceinformer.Get(ctx)

	r := &reconciler{
		ksvcr:      srcreconciler.NewKServiceReconciler(kserviceclient.Get(ctx), ksvcInformer.Lister()),
		adapterCfg: adapterCfg,
	}

	impl := slacksource.NewImpl(ctx, r)

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	logging.FromContext(ctx).Info("Setting up event handlers")

	slackSourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	ksvcInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("SlackSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
