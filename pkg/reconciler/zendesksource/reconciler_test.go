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
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	fakek8sinjectionclient "knative.dev/pkg/client/injection/kube/client/fake"

	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"
	fakeservinginjectionclient "knative.dev/serving/pkg/client/injection/client/fake"

	"github.com/triggermesh/knative-sources/pkg/apis/sources"
	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	fakeinjectionclient "github.com/triggermesh/knative-sources/pkg/client/generated/injection/client/fake"
	reconcilerv1alpha1 "github.com/triggermesh/knative-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/zendesksource"
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

// reconcilerCtor returns a Ctor for a source Reconciler.
func reconcilerCtor(cfg *adapterConfig) Ctor {
	return func(t *testing.T, ctx context.Context, ls *Listers) controller.Reconciler {
		base := common.GenericServiceReconciler{
			SinkResolver: resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
			Lister:       ls.GetServiceLister().Services,
			Client:       fakeservinginjectionclient.Get(ctx).ServingV1().Services,
		}

		r := &Reconciler{
			base:         base,
			secretClient: fakek8sinjectionclient.Get(ctx).CoreV1().Secrets,
			adapterCfg:   cfg,
		}

		return reconcilerv1alpha1.NewReconciler(ctx, logging.FromContext(ctx),
			fakeinjectionclient.Get(ctx), ls.GetZendeskSourceLister(),
			controller.GetEventRecorder(ctx), r)
	}
}

// newEventSource returns a test source object with a minimal set of pre-filled attributes.
func newEventSource() *v1alpha1.ZendeskSource {
	src := &v1alpha1.ZendeskSource{
		Spec: v1alpha1.ZendeskSourceSpec{
			Subdomain: "test",
			Email:     "test@example.com",
			Token: v1alpha1.SecretValueFromSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "keyId",
				},
			},
			WebhookUsername: "test",
			WebhookPassword: v1alpha1.SecretValueFromSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "keyId",
				},
			},
		},
	}

	// assume finalizer is already set to prevent the generated reconciler
	// from generating an extra Patch action
	src.Finalizers = []string{sources.ZendeskSourceResource.String()}

	Populate(src)

	return src
}
