// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TriggersSpec defines the desired state of the Triggers component.
type TriggersSpec struct {
	// Enable controls whether the triggers component is deployed.
	// Triggers are only deployed when Enable is explicitly set to true.
	// +optional
	Enable *bool `json:"enable,omitempty"`
}

// ShipwrightBuildSpec defines the configuration of a Shipwright Build deployment.
type ShipwrightBuildSpec struct {
	// TargetNamespace is the target namespace where Shipwright's build controller will be deployed.
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// Triggers configures the deployment of the Shipwright Triggers component.
	// When omitted, triggers are not deployed.
	// +optional
	Triggers *TriggersSpec `json:"triggers,omitempty"`
}

// TriggersEnabled returns true if the Triggers component should be deployed.
// Triggers are only deployed when spec.triggers.enable is explicitly set to true.
func (s *ShipwrightBuildSpec) TriggersEnabled() bool {
	if s.Triggers == nil || s.Triggers.Enable == nil {
		return false
	}
	return *s.Triggers.Enable
}

// ShipwrightBuildStatus defines the observed state of ShipwrightBuild
type ShipwrightBuildStatus struct {
	// Conditions holds the latest available observations of a resource's current state.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// ShipwrightBuild represents the deployment of Shipwright's build controller on a Kubernetes cluster.
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

	Items []ShipwrightBuild `json:"items"`
}

// init registers the current Schema on the Scheme Builder during initialization.
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
