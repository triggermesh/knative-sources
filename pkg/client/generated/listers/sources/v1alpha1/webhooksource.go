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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// WebhookSourceLister helps list WebhookSources.
type WebhookSourceLister interface {
	// List lists all WebhookSources in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.WebhookSource, err error)
	// WebhookSources returns an object that can list and get WebhookSources.
	WebhookSources(namespace string) WebhookSourceNamespaceLister
	WebhookSourceListerExpansion
}

// webhookSourceLister implements the WebhookSourceLister interface.
type webhookSourceLister struct {
	indexer cache.Indexer
}

// NewWebhookSourceLister returns a new WebhookSourceLister.
func NewWebhookSourceLister(indexer cache.Indexer) WebhookSourceLister {
	return &webhookSourceLister{indexer: indexer}
}

// List lists all WebhookSources in the indexer.
func (s *webhookSourceLister) List(selector labels.Selector) (ret []*v1alpha1.WebhookSource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WebhookSource))
	})
	return ret, err
}

// WebhookSources returns an object that can list and get WebhookSources.
func (s *webhookSourceLister) WebhookSources(namespace string) WebhookSourceNamespaceLister {
	return webhookSourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// WebhookSourceNamespaceLister helps list and get WebhookSources.
type WebhookSourceNamespaceLister interface {
	// List lists all WebhookSources in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.WebhookSource, err error)
	// Get retrieves the WebhookSource from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.WebhookSource, error)
	WebhookSourceNamespaceListerExpansion
}

// webhookSourceNamespaceLister implements the WebhookSourceNamespaceLister
// interface.
type webhookSourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all WebhookSources in the indexer for a given namespace.
func (s webhookSourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.WebhookSource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WebhookSource))
	})
	return ret, err
}

// Get retrieves the WebhookSource from the indexer for a given namespace and name.
func (s webhookSourceNamespaceLister) Get(name string) (*v1alpha1.WebhookSource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("webhooksource"), name)
	}
	return obj.(*v1alpha1.WebhookSource), nil
}
