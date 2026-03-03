package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ShareHolderSpec defines the desired state of ShareHolder.
type ShareHolderSpec struct {
	// Type identifies the kind of entity holding shares.
	// +kubebuilder:validation:Enum=Node;Pod;External
	Type string `json:"type"`
}

// ShareHolderStatus defines the observed state of ShareHolder.
type ShareHolderStatus struct {
	// Ready indicates the holder can participate in share operations.
	Ready bool `json:"ready,omitempty"`

	// SharesHeld is the number of shares currently stored by this holder.
	SharesHeld int `json:"sharesHeld,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Shares",type=integer,JSONPath=`.status.sharesHeld`

// ShareHolder is the Schema for the shareholders API.
type ShareHolder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShareHolderSpec   `json:"spec,omitempty"`
	Status ShareHolderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ShareHolderList contains a list of ShareHolder.
type ShareHolderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShareHolder `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShareHolder{}, &ShareHolderList{})
}
