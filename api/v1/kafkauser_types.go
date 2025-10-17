// api/v1/kafkauser_types.go
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaUser describes a user in Kafka
type KafkaUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaUserSpec   `json:"spec,omitempty"`
	Status KafkaUserStatus `json:"status,omitempty"`
}

// KafkaUserSpec defines the desired state
type KafkaUserSpec struct {
	Topic       string `json:"topic"`
	Permissions string `json:"permissions"` // read, write, admin
}

// KafkaUserStatus defines the observed state
type KafkaUserStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaUserList is a list of KafkaUser
type KafkaUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaUser `json:"items"`
}

// Ensure that KafkaUser implements runtime.Object
var _ runtime.Object = &KafkaUser{}

// Ensure that KafkaUserList implements runtime.Object
var _ runtime.Object = &KafkaUserList{}
