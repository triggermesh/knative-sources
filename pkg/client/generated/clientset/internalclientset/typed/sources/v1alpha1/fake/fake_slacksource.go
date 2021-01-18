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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/triggermesh/knative-sources/pkg/apis/sources/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSlackSources implements SlackSourceInterface
type FakeSlackSources struct {
	Fake *FakeSourcesV1alpha1
	ns   string
}

var slacksourcesResource = schema.GroupVersionResource{Group: "sources.triggermesh.io", Version: "v1alpha1", Resource: "slacksources"}

var slacksourcesKind = schema.GroupVersionKind{Group: "sources.triggermesh.io", Version: "v1alpha1", Kind: "SlackSource"}

// Get takes name of the slackSource, and returns the corresponding slackSource object, and an error if there is any.
func (c *FakeSlackSources) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.SlackSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(slacksourcesResource, c.ns, name), &v1alpha1.SlackSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlackSource), err
}

// List takes label and field selectors, and returns the list of SlackSources that match those selectors.
func (c *FakeSlackSources) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SlackSourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(slacksourcesResource, slacksourcesKind, c.ns, opts), &v1alpha1.SlackSourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SlackSourceList{ListMeta: obj.(*v1alpha1.SlackSourceList).ListMeta}
	for _, item := range obj.(*v1alpha1.SlackSourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested slackSources.
func (c *FakeSlackSources) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(slacksourcesResource, c.ns, opts))

}

// Create takes the representation of a slackSource and creates it.  Returns the server's representation of the slackSource, and an error, if there is any.
func (c *FakeSlackSources) Create(ctx context.Context, slackSource *v1alpha1.SlackSource, opts v1.CreateOptions) (result *v1alpha1.SlackSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(slacksourcesResource, c.ns, slackSource), &v1alpha1.SlackSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlackSource), err
}

// Update takes the representation of a slackSource and updates it. Returns the server's representation of the slackSource, and an error, if there is any.
func (c *FakeSlackSources) Update(ctx context.Context, slackSource *v1alpha1.SlackSource, opts v1.UpdateOptions) (result *v1alpha1.SlackSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(slacksourcesResource, c.ns, slackSource), &v1alpha1.SlackSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlackSource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSlackSources) UpdateStatus(ctx context.Context, slackSource *v1alpha1.SlackSource, opts v1.UpdateOptions) (*v1alpha1.SlackSource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(slacksourcesResource, "status", c.ns, slackSource), &v1alpha1.SlackSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlackSource), err
}

// Delete takes name of the slackSource and deletes it. Returns an error if one occurs.
func (c *FakeSlackSources) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(slacksourcesResource, c.ns, name), &v1alpha1.SlackSource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSlackSources) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(slacksourcesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SlackSourceList{})
	return err
}

// Patch applies the patch and returns the patched slackSource.
func (c *FakeSlackSources) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SlackSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(slacksourcesResource, c.ns, name, pt, data, subresources...), &v1alpha1.SlackSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SlackSource), err
}
