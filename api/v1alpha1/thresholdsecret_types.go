package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ThresholdSecretSpec defines the desired state of ThresholdSecret.
type ThresholdSecretSpec struct {
	// Threshold is the minimum number of shares required to reconstruct the secret (k).
	// +kubebuilder:validation:Minimum=2
	Threshold int `json:"threshold"`

	// Shares is the total number of shares to generate (n).
	// +kubebuilder:validation:Minimum=2
	Shares int `json:"shares"`

	// Scheme is the secret sharing scheme to use.
	// +kubebuilder:validation:Enum=sss;feldman;pedersen
	// +kubebuilder:default=feldman
	Scheme string `json:"scheme,omitempty"`

	// SecretRef references the Kubernetes Secret to protect.
	SecretRef SecretReference `json:"secretRef"`
}

// SecretReference identifies a Kubernetes Secret.
type SecretReference struct {
	// Name of the Secret.
	Name string `json:"name"`
}

// ThresholdSecretPhase represents the current lifecycle phase.
// +kubebuilder:validation:Enum=Pending;Splitting;Active;Refreshing;Degraded;Failed
type ThresholdSecretPhase string

const (
	PhasePending    ThresholdSecretPhase = "Pending"
	PhaseSplitting  ThresholdSecretPhase = "Splitting"
	PhaseActive     ThresholdSecretPhase = "Active"
	PhaseRefreshing ThresholdSecretPhase = "Refreshing"
	PhaseDegraded   ThresholdSecretPhase = "Degraded"
	PhaseFailed     ThresholdSecretPhase = "Failed"
)

// ThresholdSecretStatus defines the observed state of ThresholdSecret.
type ThresholdSecretStatus struct {
	// Phase is the current lifecycle phase.
	Phase ThresholdSecretPhase `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Threshold",type=integer,JSONPath=`.spec.threshold`
// +kubebuilder:printcolumn:name="Shares",type=integer,JSONPath=`.spec.shares`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// ThresholdSecret is the Schema for the thresholdsecrets API.
type ThresholdSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ThresholdSecretSpec   `json:"spec,omitempty"`
	Status ThresholdSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ThresholdSecretList contains a list of ThresholdSecret.
type ThresholdSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ThresholdSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ThresholdSecret{}, &ThresholdSecretList{})
}
