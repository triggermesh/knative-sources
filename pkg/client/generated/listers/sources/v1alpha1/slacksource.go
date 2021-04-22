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

// SlackSourceLister helps list SlackSources.
// All objects returned here must be treated as read-only.
type SlackSourceLister interface {
	// List lists all SlackSources in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SlackSource, err error)
	// SlackSources returns an object that can list and get SlackSources.
	SlackSources(namespace string) SlackSourceNamespaceLister
	SlackSourceListerExpansion
}

// slackSourceLister implements the SlackSourceLister interface.
type slackSourceLister struct {
	indexer cache.Indexer
}

// NewSlackSourceLister returns a new SlackSourceLister.
func NewSlackSourceLister(indexer cache.Indexer) SlackSourceLister {
	return &slackSourceLister{indexer: indexer}
}

// List lists all SlackSources in the indexer.
func (s *slackSourceLister) List(selector labels.Selector) (ret []*v1alpha1.SlackSource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SlackSource))
	})
	return ret, err
}

// SlackSources returns an object that can list and get SlackSources.
func (s *slackSourceLister) SlackSources(namespace string) SlackSourceNamespaceLister {
	return slackSourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SlackSourceNamespaceLister helps list and get SlackSources.
// All objects returned here must be treated as read-only.
type SlackSourceNamespaceLister interface {
	// List lists all SlackSources in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SlackSource, err error)
	// Get retrieves the SlackSource from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.SlackSource, error)
	SlackSourceNamespaceListerExpansion
}

// slackSourceNamespaceLister implements the SlackSourceNamespaceLister
// interface.
type slackSourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SlackSources in the indexer for a given namespace.
func (s slackSourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.SlackSource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SlackSource))
	})
	return ret, err
}

// Get retrieves the SlackSource from the indexer for a given namespace and name.
func (s slackSourceNamespaceLister) Get(name string) (*v1alpha1.SlackSource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("slacksource"), name)
	}
	return obj.(*v1alpha1.SlackSource), nil
}
