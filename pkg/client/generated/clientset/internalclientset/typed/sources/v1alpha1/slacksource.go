/*
Copyright (c) 2020 TriggerMesh, Inc

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	scheme "github.com/triggermesh/knative-sources/pkg/client/generated/clientset/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SlackSourcesGetter has a method to return a SlackSourceInterface.
// A group's client should implement this interface.
type SlackSourcesGetter interface {
	SlackSources(namespace string) SlackSourceInterface
}

// SlackSourceInterface has methods to work with SlackSource resources.
type SlackSourceInterface interface {
	Create(*v1alpha1.SlackSource) (*v1alpha1.SlackSource, error)
	Update(*v1alpha1.SlackSource) (*v1alpha1.SlackSource, error)
	UpdateStatus(*v1alpha1.SlackSource) (*v1alpha1.SlackSource, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.SlackSource, error)
	List(opts v1.ListOptions) (*v1alpha1.SlackSourceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SlackSource, err error)
	SlackSourceExpansion
}

// slackSources implements SlackSourceInterface
type slackSources struct {
	client rest.Interface
	ns     string
}

// newSlackSources returns a SlackSources
func newSlackSources(c *SourcesV1alpha1Client, namespace string) *slackSources {
	return &slackSources{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the slackSource, and returns the corresponding slackSource object, and an error if there is any.
func (c *slackSources) Get(name string, options v1.GetOptions) (result *v1alpha1.SlackSource, err error) {
	result = &v1alpha1.SlackSource{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("slacksources").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SlackSources that match those selectors.
func (c *slackSources) List(opts v1.ListOptions) (result *v1alpha1.SlackSourceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SlackSourceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("slacksources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested slackSources.
func (c *slackSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("slacksources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a slackSource and creates it.  Returns the server's representation of the slackSource, and an error, if there is any.
func (c *slackSources) Create(slackSource *v1alpha1.SlackSource) (result *v1alpha1.SlackSource, err error) {
	result = &v1alpha1.SlackSource{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("slacksources").
		Body(slackSource).
		Do().
		Into(result)
	return
}

// Update takes the representation of a slackSource and updates it. Returns the server's representation of the slackSource, and an error, if there is any.
func (c *slackSources) Update(slackSource *v1alpha1.SlackSource) (result *v1alpha1.SlackSource, err error) {
	result = &v1alpha1.SlackSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("slacksources").
		Name(slackSource.Name).
		Body(slackSource).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *slackSources) UpdateStatus(slackSource *v1alpha1.SlackSource) (result *v1alpha1.SlackSource, err error) {
	result = &v1alpha1.SlackSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("slacksources").
		Name(slackSource.Name).
		SubResource("status").
		Body(slackSource).
		Do().
		Into(result)
	return
}

// Delete takes name of the slackSource and deletes it. Returns an error if one occurs.
func (c *slackSources) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("slacksources").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *slackSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("slacksources").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched slackSource.
func (c *slackSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SlackSource, err error) {
	result = &v1alpha1.SlackSource{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("slacksources").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}