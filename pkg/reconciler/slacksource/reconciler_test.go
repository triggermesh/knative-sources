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

package slacksource

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"
	fakeservinginjectionclient "knative.dev/serving/pkg/client/injection/client/fake"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	fakeinjectionclient "github.com/triggermesh/knative-sources/pkg/client/generated/injection/client/fake"
	reconcilerv1alpha1 "github.com/triggermesh/knative-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/slacksource"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common"
	. "github.com/triggermesh/knative-sources/pkg/reconciler/testing"
)

func TestReconcileSource(t *testing.T) {
	adapterCfg := &adapterConfig{
		Image:   "registry/image:tag",
		configs: &source.EmptyVarsGenerator{},
	}

	var (
		ctor      = reconcilerCtor(adapterCfg)
		src       = newEventSource()
		adapterFn = adapterServiceBuilder(src, adapterCfg)
	)

	TestReconcile(t, ctor, src, adapterFn)
}

// reconcilerCtor returns a Ctor for a SlackSource Reconciler.
func reconcilerCtor(cfg *adapterConfig) Ctor {
	return func(t *testing.T, ctx context.Context, ls *Listers) controller.Reconciler {
		base := common.GenericServiceReconciler{
			SinkResolver: resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
			Lister:       ls.GetServiceLister().Services,
			Client:       fakeservinginjectionclient.Get(ctx).ServingV1().Services,
		}

		r := &Reconciler{
			base:       base,
			adapterCfg: cfg,
		}

		return reconcilerv1alpha1.NewReconciler(ctx, logging.FromContext(ctx),
			fakeinjectionclient.Get(ctx), ls.GetSlackSourceLister(),
			controller.GetEventRecorder(ctx), r)
	}
}

// newEventSource returns a test source object with a minimal set of pre-filled attributes.
func newEventSource() *v1alpha1.SlackSource {
	src := &v1alpha1.SlackSource{
		Spec: v1alpha1.SlackSourceSpec{},
	}

	Populate(src)

	return src
}
