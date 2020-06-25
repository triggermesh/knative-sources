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

package testing

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	sourcesv1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	fakeeventingclient "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	sourceslistersv1alpha2 "knative.dev/eventing/pkg/client/listers/sources/v1alpha2"

	rt "knative.dev/pkg/reconciler/testing"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	fakeservingclient "knative.dev/serving/pkg/client/clientset/versioned/fake"
	servinglistersv1 "knative.dev/serving/pkg/client/listers/serving/v1"

	"github.com/triggermesh/knative-sources/slack/pkg/apis/sources/v1alpha1"
	fakeclient "github.com/triggermesh/knative-sources/slack/pkg/client/generated/clientset/internalclientset/fake"
	listersv1alpha1 "github.com/triggermesh/knative-sources/slack/pkg/client/generated/listers/sources/v1alpha1"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakeclient.AddToScheme,
	fakeservingclient.AddToScheme,
	fakeeventingclient.AddToScheme,
}

// NewScheme returns a new scheme populated with the types defined in clientSetSchemes.
func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	sb := runtime.NewSchemeBuilder(clientSetSchemes...)
	if err := sb.AddToScheme(scheme); err != nil {
		panic(fmt.Errorf("error building Scheme: %s", err))
	}

	return scheme
}

// Listers returns listers and objects filtered from those listers.
type Listers struct {
	sorter rt.ObjectSorter
}

// NewListers returns a new instance of Listers initialized with the given objects.
func NewListers(scheme *runtime.Scheme, objs []runtime.Object) Listers {
	ls := Listers{
		sorter: rt.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

// IndexerFor returns the indexer for the given object.
func (l *Listers) IndexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

// GetSlackSourceObjects returns objects from the sources API.
func (l *Listers) GetSlackSourceObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeclient.AddToScheme)
}

// GetServingObjects returns objects from the serving API.
func (l *Listers) GetServingObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeservingclient.AddToScheme)
}

// GetEventingObjects returns objects from the eventing API.
func (l *Listers) GetEventingObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeeventingclient.AddToScheme)
}

// GetSlackSourceLister returns a Lister for SlackSource objects.
func (l *Listers) GetSlackSourceLister() listersv1alpha1.SlackSourceLister {
	return listersv1alpha1.NewSlackSourceLister(l.IndexerFor(&v1alpha1.SlackSource{}))
}

// GetServiceLister returns a lister for Service objects.
func (l *Listers) GetServiceLister() servinglistersv1.ServiceLister {
	return servinglistersv1.NewServiceLister(l.IndexerFor(&servingv1.Service{}))
}

// GetEventingLister returns a lister for Service objects.
func (l *Listers) GetEventingLister() sourceslistersv1alpha2.SinkBindingLister {
	return sourceslistersv1alpha2.NewSinkBindingLister(l.IndexerFor(&sourcesv1alpha2.SinkBinding{}))
}
