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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgotesting "k8s.io/client-go/testing"

	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	knlog "knative.dev/pkg/logging/testing"
	rt "knative.dev/pkg/reconciler/testing"
	"knative.dev/pkg/resolver"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	fakeservinginjectionclient "knative.dev/serving/pkg/client/injection/client/fake"

	"github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	fakeinjectionclient "github.com/triggermesh/knative-sources/pkg/client/generated/injection/client/fake"
	reconcilerv1alpha1 "github.com/triggermesh/knative-sources/pkg/client/generated/injection/reconciler/sources/v1alpha1/slacksource"
	srcreconciler "github.com/triggermesh/knative-sources/pkg/reconciler"
	"github.com/triggermesh/knative-sources/pkg/reconciler/resources"
	st "github.com/triggermesh/knative-sources/slack/pkg/reconciler/testing"
)

const (
	tNs   = "testns"
	tName = "test"
	tKey  = tNs + "/" + tName
	tUID  = types.UID("00000000-0000-0000-0000-000000000000")

	tSvcName = adapterName + "-" + tName
	tImg     = "registry/image:tag"
)

var (
	tAdapterCfg = &adapterConfig{
		obsConfig: &source.EmptyVarsGenerator{},
		Image:     st.Image,
	}
	tSinkURI = &apis.URL{
		Scheme: "http",
		Host:   "sink.example.com",
		Path:   "/",
	}
	tAdapterURI = &apis.URL{
		Scheme: "http",
		Host:   "adapter.example.com",
		Path:   "/",
	}
)

// Test the Reconcile method of the controller.Reconciler implemented by our controller.
//
// The environment for each test case is set up as follows:
//  1. MakeFactory initializes fake clients with the objects declared in the test case
//  2. MakeFactory injects those clients into a context along with fake event recorders, etc.
//  3. A Reconciler is constructed via a Ctor function using the values injected above
//  4. The Reconciler returned by MakeFactory is used to run the test case
func TestReconcile(t *testing.T) {
	testCases := rt.TableTest{
		// Creation

		{
			Name: "Slack source creation",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				makeSlackSource(),
			},
			WantCreates: []runtime.Object{
				makeSlackAdapter(),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: makeSlackSource(
					sourceWithAdapter(makeSlackAdapter()),
					sourceWithSink,
					sourceWithCloudEventsAttr,
				),
			}},
			WantEvents: []string{
				createAdapterEvent(),
			},
		},

		// Deletion

		{
			Name: "Slack source deletion",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),
				makeSlackAdapter(),
				makeSlackSource(
					sourceWithAdapter(makeSlackAdapter()),
					sourceWithSink,
					sourceWithCloudEventsAttr,
					sourceWithDeletionTimestamp,
				),
			},
		},

		// Lifecycle

		{
			Name: "Adapter becomes Ready",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),

				makeSlackSource(
					sourceWithAdapter(makeSlackAdapter()),
					sourceWithSink,
					sourceWithCloudEventsAttr,
				),
				makeSlackAdapter(
					withAdapterStatus(corev1.ConditionTrue),
					withAdapterAddress(tAdapterURI),
				),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: makeSlackSource(
					sourceWithAdapter(
						makeSlackAdapter(
							withAdapterStatus(corev1.ConditionTrue),
							withAdapterAddress(tAdapterURI),
						),
					),
					sourceWithSink,
					sourceWithCloudEventsAttr,
				),
			}},
		},

		{
			Name: "Adapter becomes Not Ready",
			Key:  tKey,
			Objects: []runtime.Object{
				newAdressable(),

				makeSlackSource(
					sourceWithAdapter(makeSlackAdapter(
						withAdapterStatus(corev1.ConditionTrue),
						withAdapterAddress(tAdapterURI),
					)),
					sourceWithSink,
					sourceWithCloudEventsAttr,
				),
				makeSlackAdapter(
					withAdapterStatus(corev1.ConditionFalse),
					withAdapterAddress(tAdapterURI),
				),
			},
			WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
				Object: makeSlackSource(
					sourceWithAdapter(
						makeSlackAdapter(
							withAdapterStatus(corev1.ConditionFalse),
							withAdapterAddress(tAdapterURI),
						),
					),
					sourceWithSink,
					sourceWithCloudEventsAttr,
				),
			}},
		},
	}

	logger := knlog.TestLogger(t)
	testCases.Test(t, st.MakeFactory(reconcilerCtor, logger))
}

// reconcilerCtor returns a Ctor for a SlackSource Reconciler.
func reconcilerCtor(ctx context.Context, ls *st.Listers) controller.Reconciler {
	logger := logging.FromContext(ctx)

	r := &reconciler{
		ksvcr:        srcreconciler.NewKServiceReconciler(fakeservinginjectionclient.Get(ctx), ls.GetServiceLister()),
		sinkResolver: resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
		adapterCfg:   tAdapterCfg,
	}

	return reconcilerv1alpha1.NewReconciler(ctx, logger, fakeinjectionclient.Get(ctx), ls.GetSlackSourceLister(), controller.GetEventRecorder(ctx), r)
}

// Slack Source

type sourceOpt func(*v1alpha1.SlackSource)

func makeSlackSource(opts ...sourceOpt) *v1alpha1.SlackSource {
	addr := newAdressable()
	addrGVK := addr.GetGroupVersionKind()

	o := &v1alpha1.SlackSource{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SlackSource",
			APIVersion: "sources.triggermesh.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
			UID:       tUID,
		},
		Spec: v1alpha1.SlackSourceSpec{
			SourceSpec: duckv1.SourceSpec{
				Sink: duckv1.Destination{
					Ref: &duckv1.KReference{
						APIVersion: addrGVK.GroupVersion().String(),
						Kind:       addrGVK.Kind,
						Name:       addr.Name,
					},
				},
			},
		},
	}

	o.Status.InitializeConditions()

	for _, f := range opts {
		f(o)
	}
	return o
}

func sourceWithAdapter(s *servingv1.Service) sourceOpt {
	return func(ss *v1alpha1.SlackSource) {
		ss.Status.PropagateAvailability(s)
	}
}

func sourceWithSink(s *v1alpha1.SlackSource) {
	s.Status.MarkSink(tSinkURI)
}

func sourceWithCloudEventsAttr(s *v1alpha1.SlackSource) {
	s.Status.CloudEventAttributes = []duckv1.CloudEventAttributes{{Type: v1alpha1.SlackSourceEventType}}
}

// deleted marks the source as deleted.
func sourceWithDeletionTimestamp(src *v1alpha1.SlackSource) {
	t := metav1.Unix(0, 0)
	src.SetDeletionTimestamp(&t)
}

// Slack Adapter

type adapterOpt func(*servingv1.Service)

// makeSlackAdapter returns a test Service object with pre-filled attributes.
func makeSlackAdapter(opts ...adapterOpt) *servingv1.Service {
	name := kmeta.ChildName(adapterName+"-", st.Name)
	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      name,
			Labels: map[string]string{
				resources.AppNameLabel:      adapterName,
				resources.AppInstanceLabel:  st.Name,
				resources.AppComponentLabel: resources.AdapterComponent,
				resources.AppPartOfLabel:    partOf,
				resources.AppManagedByLabel: managedBy,
			}, OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(st.NewOwnerRefable(
					tName,
					(&v1alpha1.SlackSource{}).GetGroupVersionKind(),
					tUID,
				)),
			},
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							resources.AppNameLabel:      adapterName,
							resources.AppInstanceLabel:  st.Name,
							resources.AppComponentLabel: resources.AdapterComponent,
							resources.AppPartOfLabel:    partOf,
							resources.AppManagedByLabel: managedBy,
						},
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "adapter", // defaulted by resource package
								Image: tImg,
								Env: []corev1.EnvVar{
									{
										Name:  "NAMESPACE",
										Value: tNs,
									},
									{
										Name:  "NAME",
										Value: name,
									},
									{
										Name:  "K_SINK",
										Value: tSinkURI.String(),
									},
									{
										Name: source.EnvLoggingCfg,
									}, {
										Name: source.EnvMetricsCfg,
									}, {
										Name: source.EnvTracingCfg,
									},
								},
							}},
						},
					},
				},
			},
		},
	}

	for _, f := range opts {
		f(ksvc)
	}
	return ksvc
}

func withAdapterStatus(status corev1.ConditionStatus) adapterOpt {
	return func(ksvc *servingv1.Service) {
		ksvc.Status.SetConditions(apis.Conditions{{
			Type:   v1alpha1.ConditionReady,
			Status: status,
		}})
	}
}

func withAdapterAddress(url *apis.URL) adapterOpt {
	return func(ksvc *servingv1.Service) {
		ksvc.Status.URL = url
	}
}

// Addressable for Source

// newAdressable returns a test Addressable to be used as a sink.
func newAdressable() *eventingv1beta1.Broker {
	return &eventingv1beta1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
		},
		Status: eventingv1beta1.BrokerStatus{
			Address: duckv1.Addressable{
				URL: tSinkURI,
			},
		},
	}
}

// Events

func createAdapterEvent() string {
	return Eventf(corev1.EventTypeNormal, srcreconciler.ReasonAdapterCreate, "Created knative service: \"%s/%s\"",
		tNs, tSvcName)
}

// Eventf returns the attributes of an API event in the format returned by
// Kubernetes' FakeRecorder.
func Eventf(eventtype, reason, messageFmt string, args ...interface{}) string {
	return fmt.Sprintf(eventtype+" "+reason+" "+messageFmt, args...)
}
