//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022 Quentin Lemaire <quentin@lemairepro.fr>.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthSecretReference) DeepCopyInto(out *AuthSecretReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthSecretReference.
func (in *AuthSecretReference) DeepCopy() *AuthSecretReference {
	if in == nil {
		return nil
	}
	out := new(AuthSecretReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Feed) DeepCopyInto(out *Feed) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Feed.
func (in *Feed) DeepCopy() *Feed {
	if in == nil {
		return nil
	}
	out := new(Feed)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Feed) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FeedList) DeepCopyInto(out *FeedList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Feed, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FeedList.
func (in *FeedList) DeepCopy() *FeedList {
	if in == nil {
		return nil
	}
	out := new(FeedList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FeedList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FeedSpec) DeepCopyInto(out *FeedSpec) {
	*out = *in
	if in.ParentDirID != nil {
		in, out := &in.ParentDirID, &out.ParentDirID
		*out = new(uint)
		**out = **in
	}
	if in.DeleteOldFiles != nil {
		in, out := &in.DeleteOldFiles, &out.DeleteOldFiles
		*out = new(bool)
		**out = **in
	}
	if in.DontProcessWholeFeed != nil {
		in, out := &in.DontProcessWholeFeed, &out.DontProcessWholeFeed
		*out = new(bool)
		**out = **in
	}
	if in.Paused != nil {
		in, out := &in.Paused, &out.Paused
		*out = new(bool)
		**out = **in
	}
	out.AuthSecretRef = in.AuthSecretRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FeedSpec.
func (in *FeedSpec) DeepCopy() *FeedSpec {
	if in == nil {
		return nil
	}
	out := new(FeedSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FeedStatus) DeepCopyInto(out *FeedStatus) {
	*out = *in
	if in.ID != nil {
		in, out := &in.ID, &out.ID
		*out = new(uint)
		**out = **in
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FeedStatus.
func (in *FeedStatus) DeepCopy() *FeedStatus {
	if in == nil {
		return nil
	}
	out := new(FeedStatus)
	in.DeepCopyInto(out)
	return out
}
