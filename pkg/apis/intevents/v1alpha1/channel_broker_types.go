/*
Copyright 2020 Google LLC

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

package v1alpha1

import (
	"github.com/google/knative-gcp/pkg/apis/broker/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChannelBroker manages the set of data plane components servicing
// one or more Broker objects and their associated Triggers.
type ChannelBroker struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the ChannelBroker.
	Spec ChannelBrokerSpec `json:"spec,omitempty"`

	// Status represents the current state of the ChannelBroker. This data may be out of
	// date.
	// +optional
	Status ChannelBrokerStatus `json:"status,omitempty"`
}

var (
	// Check that ChannelBroker can be validated, can be defaulted, and has immutable fields.
	_ apis.Validatable = (*ChannelBroker)(nil)
	_ apis.Defaultable = (*ChannelBroker)(nil)

	// Check that ChannelBroker can return its spec untyped.
	_ apis.HasSpec = (*ChannelBroker)(nil)

	_ runtime.Object = (*ChannelBroker)(nil)

	// Check that we can create OwnerReferences to a ChannelBroker.
	_ kmeta.OwnerRefable = (*ChannelBroker)(nil)

	// Check that ChannelBroker implements the KRShaped duck type.
	_ duckv1.KRShaped = (*ChannelBroker)(nil)
)

// ChannelBrokerSpec defines the desired state of a Brokercell.
type ChannelBrokerSpec struct {
	eventingv1beta1.BrokerSpec `json:",inline"`
}

// ChannelBrokerStatus represents the current state of a ChannelBroker.
type ChannelBrokerStatus struct {
	v1beta1.BrokerStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChannelBrokerList is a collection of ChannelBrokers.
type ChannelBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ChannelBroker `json:"items"`
}

// GetGroupVersionKind returns GroupVersionKind for Brokers
func (bc *ChannelBroker) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ChannelBroker")
}

// GetUntypedSpec returns the spec of the ChannelBroker.
func (bc *ChannelBroker) GetUntypedSpec() interface{} {
	return bc.Spec
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*ChannelBroker) GetConditionSet() apis.ConditionSet {
	return brokerCellCondSet
}

// GetStatus retrieves the status of the ChannelBroker. Implements the KRShaped interface.
func (bc *ChannelBroker) GetStatus() *duckv1.Status {
	return &bc.Status.Status
}
