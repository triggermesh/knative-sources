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

// FakeOciMetricsSources implements OciMetricsSourceInterface
type FakeOciMetricsSources struct {
	Fake *FakeSourcesV1alpha1
	ns   string
}

var ocimetricssourcesResource = schema.GroupVersionResource{Group: "sources.triggermesh.io", Version: "v1alpha1", Resource: "ocimetricssources"}

var ocimetricssourcesKind = schema.GroupVersionKind{Group: "sources.triggermesh.io", Version: "v1alpha1", Kind: "OciMetricsSource"}

// Get takes name of the ociMetricsSource, and returns the corresponding ociMetricsSource object, and an error if there is any.
func (c *FakeOciMetricsSources) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.OciMetricsSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ocimetricssourcesResource, c.ns, name), &v1alpha1.OciMetricsSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OciMetricsSource), err
}

// List takes label and field selectors, and returns the list of OciMetricsSources that match those selectors.
func (c *FakeOciMetricsSources) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.OciMetricsSourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ocimetricssourcesResource, ocimetricssourcesKind, c.ns, opts), &v1alpha1.OciMetricsSourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.OciMetricsSourceList{ListMeta: obj.(*v1alpha1.OciMetricsSourceList).ListMeta}
	for _, item := range obj.(*v1alpha1.OciMetricsSourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested ociMetricsSources.
func (c *FakeOciMetricsSources) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ocimetricssourcesResource, c.ns, opts))

}

// Create takes the representation of a ociMetricsSource and creates it.  Returns the server's representation of the ociMetricsSource, and an error, if there is any.
func (c *FakeOciMetricsSources) Create(ctx context.Context, ociMetricsSource *v1alpha1.OciMetricsSource, opts v1.CreateOptions) (result *v1alpha1.OciMetricsSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ocimetricssourcesResource, c.ns, ociMetricsSource), &v1alpha1.OciMetricsSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OciMetricsSource), err
}

// Update takes the representation of a ociMetricsSource and updates it. Returns the server's representation of the ociMetricsSource, and an error, if there is any.
func (c *FakeOciMetricsSources) Update(ctx context.Context, ociMetricsSource *v1alpha1.OciMetricsSource, opts v1.UpdateOptions) (result *v1alpha1.OciMetricsSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ocimetricssourcesResource, c.ns, ociMetricsSource), &v1alpha1.OciMetricsSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OciMetricsSource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeOciMetricsSources) UpdateStatus(ctx context.Context, ociMetricsSource *v1alpha1.OciMetricsSource, opts v1.UpdateOptions) (*v1alpha1.OciMetricsSource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(ocimetricssourcesResource, "status", c.ns, ociMetricsSource), &v1alpha1.OciMetricsSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OciMetricsSource), err
}

// Delete takes name of the ociMetricsSource and deletes it. Returns an error if one occurs.
func (c *FakeOciMetricsSources) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(ocimetricssourcesResource, c.ns, name), &v1alpha1.OciMetricsSource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeOciMetricsSources) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ocimetricssourcesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.OciMetricsSourceList{})
	return err
}

// Patch applies the patch and returns the patched ociMetricsSource.
func (c *FakeOciMetricsSources) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.OciMetricsSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ocimetricssourcesResource, c.ns, name, pt, data, subresources...), &v1alpha1.OciMetricsSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OciMetricsSource), err
}
