package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ClusterID",type=string,JSONPath=`.status.klusterID`
// +kubebuilder:printcolumn:name="Progress",type=string,JSONPath=`.status.progress`
type Kluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              KlusterSpec   `json:"spec"`
	Status            KlusterStatus `json:"status,omitempty"`
}

// +kubebuilder:validation:Required
type KlusterSpec struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Region      string     `json:"region"`
	TokenSecret string     `json:"tokensecret"`
	Nodepools   []NodePool `json:"nodepools"`
}

type NodePool struct {
	Size  string `json:"size"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type KlusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kluster `json:"items"`
}
type KlusterStatus struct {
	Process    string `json:"process,omitempty"`
	KlusterId  string `json:"klusterId,omitempty"`
	KubeConfig string `json:"kubeConfig,omitempty"`
}

// package v1alpha

// import (
//     metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// // +genclient
// // +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// type Kluster struct {
//     metav1.TypeMeta   `json:",inline"`
//     metav1.ObjectMeta `json:"metadata,omitempty"`
//     Spec              KlusterSpec `json:"spec"`
// }

// type KlusterSpec struct {
//     Name      string     `json:"name"`
//     Version   string     `json:"version"`
//     Region    string     `json:"region"`
//     TokenSecret string    `json:"tokensecret"`
//     Nodepools []NodePool `json:"nodepools"`
// }

// type NodePool struct {
//     Size  string `json:"size"`
//     Name  string `json:"name"`
//     Count int    `json:"count"`
// }

// // +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// type KlusterList struct {
//     metav1.TypeMeta `json:",inline"`
//     metav1.ListMeta `json:"metadata,omitempty"`
//     Items           []Kluster `json:"items"`
// }
