//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ShipwrightBuild) DeepCopyInto(out *ShipwrightBuild) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ShipwrightBuild.
func (in *ShipwrightBuild) DeepCopy() *ShipwrightBuild {
	if in == nil {
		return nil
	}
	out := new(ShipwrightBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ShipwrightBuild) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ShipwrightBuildList) DeepCopyInto(out *ShipwrightBuildList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ShipwrightBuild, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ShipwrightBuildList.
func (in *ShipwrightBuildList) DeepCopy() *ShipwrightBuildList {
	if in == nil {
		return nil
	}
	out := new(ShipwrightBuildList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ShipwrightBuildList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ShipwrightBuildSpec) DeepCopyInto(out *ShipwrightBuildSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ShipwrightBuildSpec.
func (in *ShipwrightBuildSpec) DeepCopy() *ShipwrightBuildSpec {
	if in == nil {
		return nil
	}
	out := new(ShipwrightBuildSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ShipwrightBuildStatus) DeepCopyInto(out *ShipwrightBuildStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ShipwrightBuildStatus.
func (in *ShipwrightBuildStatus) DeepCopy() *ShipwrightBuildStatus {
	if in == nil {
		return nil
	}
	out := new(ShipwrightBuildStatus)
	in.DeepCopyInto(out)
	return out
}
