// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventSourceStatus) DeepCopyInto(out *EventSourceStatus) {
	*out = *in
	in.SourceStatus.DeepCopyInto(&out.SourceStatus)
	in.AddressStatus.DeepCopyInto(&out.AddressStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventSourceStatus.
func (in *EventSourceStatus) DeepCopy() *EventSourceStatus {
	if in == nil {
		return nil
	}
	out := new(EventSourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretValueFromSource) DeepCopyInto(out *SecretValueFromSource) {
	*out = *in
	if in.SecretKeyRef != nil {
		in, out := &in.SecretKeyRef, &out.SecretKeyRef
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretValueFromSource.
func (in *SecretValueFromSource) DeepCopy() *SecretValueFromSource {
	if in == nil {
		return nil
	}
	out := new(SecretValueFromSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SlackSource) DeepCopyInto(out *SlackSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SlackSource.
func (in *SlackSource) DeepCopy() *SlackSource {
	if in == nil {
		return nil
	}
	out := new(SlackSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SlackSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SlackSourceList) DeepCopyInto(out *SlackSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SlackSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SlackSourceList.
func (in *SlackSourceList) DeepCopy() *SlackSourceList {
	if in == nil {
		return nil
	}
	out := new(SlackSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SlackSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SlackSourceSpec) DeepCopyInto(out *SlackSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	if in.SigningSecret != nil {
		in, out := &in.SigningSecret, &out.SigningSecret
		*out = new(SecretValueFromSource)
		(*in).DeepCopyInto(*out)
	}
	if in.AppID != nil {
		in, out := &in.AppID, &out.AppID
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SlackSourceSpec.
func (in *SlackSourceSpec) DeepCopy() *SlackSourceSpec {
	if in == nil {
		return nil
	}
	out := new(SlackSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueFromField) DeepCopyInto(out *ValueFromField) {
	*out = *in
	if in.ValueFromSecret != nil {
		in, out := &in.ValueFromSecret, &out.ValueFromSecret
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueFromField.
func (in *ValueFromField) DeepCopy() *ValueFromField {
	if in == nil {
		return nil
	}
	out := new(ValueFromField)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZendeskSource) DeepCopyInto(out *ZendeskSource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZendeskSource.
func (in *ZendeskSource) DeepCopy() *ZendeskSource {
	if in == nil {
		return nil
	}
	out := new(ZendeskSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ZendeskSource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZendeskSourceList) DeepCopyInto(out *ZendeskSourceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ZendeskSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZendeskSourceList.
func (in *ZendeskSourceList) DeepCopy() *ZendeskSourceList {
	if in == nil {
		return nil
	}
	out := new(ZendeskSourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ZendeskSourceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZendeskSourceSpec) DeepCopyInto(out *ZendeskSourceSpec) {
	*out = *in
	in.SourceSpec.DeepCopyInto(&out.SourceSpec)
	in.Token.DeepCopyInto(&out.Token)
	in.WebhookPassword.DeepCopyInto(&out.WebhookPassword)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZendeskSourceSpec.
func (in *ZendeskSourceSpec) DeepCopy() *ZendeskSourceSpec {
	if in == nil {
		return nil
	}
	out := new(ZendeskSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZendeskSourceStatus) DeepCopyInto(out *ZendeskSourceStatus) {
	*out = *in
	in.EventSourceStatus.DeepCopyInto(&out.EventSourceStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZendeskSourceStatus.
func (in *ZendeskSourceStatus) DeepCopy() *ZendeskSourceStatus {
	if in == nil {
		return nil
	}
	out := new(ZendeskSourceStatus)
	in.DeepCopyInto(out)
	return out
}
