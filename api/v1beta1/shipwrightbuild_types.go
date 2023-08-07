// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ShipwrightBuildSpec defines the desired state of ShipwrightBuild
type ShipwrightBuildSpec struct {
	// TargetNamespace is the target namespace where Shipwright's build controller will be deployed.
	TargetNamespace string `json:"targetNamespace,omitempty"`
}

// ShipwrightBuildStatus defines the observed state of ShipwrightBuild
type ShipwrightBuildStatus struct {
	// Conditions holds the latest available observations of a resource's current state.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ShipwrightBuild is the Schema for the shipwrightbuilds API
type ShipwrightBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShipwrightBuildSpec   `json:"spec,omitempty"`
	Status ShipwrightBuildStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ShipwrightBuildList contains a list of ShipwrightBuild
type ShipwrightBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShipwrightBuild `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShipwrightBuild{}, &ShipwrightBuildList{})
}

// IsReady returns true the Ready condition status is True
func (status ShipwrightBuildStatus) IsReady() bool {
	for _, condition := range status.Conditions {
		if condition.Type == "Ready" && condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}
