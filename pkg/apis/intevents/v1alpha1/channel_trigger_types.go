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

// ChannelTrigger manages the set of data plane components servicing
// one or more Broker objects and their associated Triggers.
type ChannelTrigger struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the ChannelTrigger.
	Spec ChannelTriggerSpec `json:"spec,omitempty"`

	// Status represents the current state of the ChannelTrigger. This data may be out of
	// date.
	// +optional
	Status ChannelTriggerStatus `json:"status,omitempty"`
}

var (
	// Check that ChannelTrigger can be validated, can be defaulted, and has immutable fields.
	_ apis.Validatable = (*ChannelTrigger)(nil)
	_ apis.Defaultable = (*ChannelTrigger)(nil)

	// Check that ChannelTrigger can return its spec untyped.
	_ apis.HasSpec = (*ChannelTrigger)(nil)

	_ runtime.Object = (*ChannelTrigger)(nil)

	// Check that we can create OwnerReferences to a ChannelTrigger.
	_ kmeta.OwnerRefable = (*ChannelTrigger)(nil)

	// Check that ChannelTrigger implements the KRShaped duck type.
	_ duckv1.KRShaped = (*ChannelTrigger)(nil)
)

// ChannelTriggerSpec defines the desired state of a Brokercell.
type ChannelTriggerSpec struct {
	eventingv1beta1.TriggerSpec `json:",inline"`
}

// ChannelTriggerStatus represents the current state of a ChannelTrigger.
type ChannelTriggerStatus struct {
	v1beta1.TriggerStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChannelTriggerList is a collection of ChannelTriggers.
type ChannelTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ChannelTrigger `json:"items"`
}

// GetGroupVersionKind returns GroupVersionKind for Brokers
func (bc *ChannelTrigger) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ChannelTrigger")
}

// GetUntypedSpec returns the spec of the ChannelTrigger.
func (bc *ChannelTrigger) GetUntypedSpec() interface{} {
	return bc.Spec
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*ChannelTrigger) GetConditionSet() apis.ConditionSet {
	return brokerCellCondSet
}

// GetStatus retrieves the status of the ChannelTrigger. Implements the KRShaped interface.
func (bc *ChannelTrigger) GetStatus() *duckv1.Status {
	return &bc.Status.Status
}
