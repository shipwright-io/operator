// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the operator v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=operator.shipwright.io
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "operator.shipwright.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(GroupVersion,
		&ShipwrightBuild{},
		&ShipwrightBuildList{},
	)
	metav1.AddToGroupVersion(s, GroupVersion)
	return nil
}
