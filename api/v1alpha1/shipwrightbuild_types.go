// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ShipwrightBuildSpec defines the desired state of ShipwrightBuild
type ShipwrightBuildSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ShipwrightBuild. Edit ShipwrightBuild_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// ShipwrightBuildStatus defines the observed state of ShipwrightBuild
type ShipwrightBuildStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ShipwrightBuild is the Schema for the shipwrightbuilds API
type ShipwrightBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShipwrightBuildSpec   `json:"spec,omitempty"`
	Status ShipwrightBuildStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ShipwrightBuildList contains a list of ShipwrightBuild
type ShipwrightBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShipwrightBuild `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShipwrightBuild{}, &ShipwrightBuildList{})
}
