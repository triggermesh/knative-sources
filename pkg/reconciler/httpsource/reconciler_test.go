/*
Copyright (c) 2020-2021 TriggerMesh Inc.

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

package httpsource

import (
	"context"
	"testing"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	rt "knative.dev/pkg/reconciler/testing"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	fakeinjectionclient "github.com/triggermesh/knative-sources/pkg/client/generated/injection/client/fake"
	reconcilerv1alpha1 "github.com/triggermesh/knative-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/httpsource"
	"github.com/triggermesh/knative-sources/pkg/reconciler/common"
	. "github.com/triggermesh/knative-sources/pkg/reconciler/testing"
)

func TestReconcileSource(t *testing.T) {
	adapterCfg := &adapterConfig{
		Image:   "registry/image:tag",
		configs: &source.EmptyVarsGenerator{},
	}

	ctor := reconcilerCtor(adapterCfg)
	src := newEventSource()
	ab := adapterBuilder(adapterCfg)

	TestReconcileAdapter(t, ctor, src, ab)
}

// reconcilerCtor returns a Ctor for a source Reconciler.
func reconcilerCtor(cfg *adapterConfig) Ctor {
	return func(t *testing.T, ctx context.Context, _ *rt.TableRow, ls *Listers) controller.Reconciler {
		r := &Reconciler{
			base:       NewTestServiceReconciler(ctx, ls),
			adapterCfg: cfg,
			srcLister:  ls.GetHTTPSourceLister().HTTPSources,
		}

		return reconcilerv1alpha1.NewReconciler(ctx, logging.FromContext(ctx),
			fakeinjectionclient.Get(ctx), ls.GetHTTPSourceLister(),
			controller.GetEventRecorder(ctx), r)
	}
}

// newEventSource returns a test source object with a minimal set of pre-filled attributes.
func newEventSource() *v1alpha1.HTTPSource {
	src := &v1alpha1.HTTPSource{
		Spec: v1alpha1.HTTPSourceSpec{
			EventType: "my.event",
		},
	}

	Populate(src)

	return src
}

// adapterBuilder returns a slim Reconciler containing only the fields accessed
// by r.BuildAdapter().
func adapterBuilder(cfg *adapterConfig) common.AdapterServiceBuilder {
	return &Reconciler{
		adapterCfg: cfg,
	}
}
