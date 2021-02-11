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

package v1beta1

import (
	"context"

	eventingduckv1beta1 "knative.dev/eventing/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"

	"github.com/google/knative-gcp/pkg/apis/configs/brokerdelivery"
)

// SetDefaults sets the default field values for a Broker.
func (b *Broker) SetDefaults(ctx context.Context) {
	// Apply the default Broker delivery settings from the context.
	withNS := apis.WithinParent(ctx, b.ObjectMeta)
	b.Spec.Delivery = GetDefaultDeliverySpec(withNS, b.Spec.Delivery)
}

// GetDefaultDeliverySpec returns the default DeliverySpec. It will take in the existing
// DeliverySpec and returns a defaulted version. Note that you need to use the returned value, you
// cannot rely on mutation of the passed in value.
// The passed in context must have set `aps.WithinParent`.
func GetDefaultDeliverySpec(ctx context.Context, deliverySpec *eventingduckv1beta1.DeliverySpec) *eventingduckv1beta1.DeliverySpec {
	deliverySpecDefaults := brokerdelivery.FromContextOrDefaults(ctx).BrokerDeliverySpecDefaults
	if deliverySpecDefaults == nil {
		// TODO This should probably either fail closed or give some sane in-code default values.
		// As-is, this will just let undefaulted things into the system, which may cause issues like
		// nil pointer dereferences later in the system.
		logging.FromContext(ctx).Error("Failed to get the BrokerDeliverySpecDefaults")
		return deliverySpec
	}
	// Set the default delivery spec.
	if deliverySpec == nil {
		deliverySpec = &eventingduckv1beta1.DeliverySpec{}
	}
	ns := apis.ParentMeta(ctx).Namespace
	if deliverySpec.BackoffPolicy == nil || deliverySpec.BackoffDelay == nil {
		// Set both defaults if one of the backoff delay or backoff policy are not specified.
		deliverySpec.BackoffPolicy = deliverySpecDefaults.BackoffPolicy(ns)
		deliverySpec.BackoffDelay = deliverySpecDefaults.BackoffDelay(ns)
	}
	if deliverySpec.DeadLetterSink == nil {
		deliverySpec.DeadLetterSink = deliverySpecDefaults.DeadLetterSink(ns)
	}
	if deliverySpec.Retry == nil && deliverySpec.DeadLetterSink != nil {
		// Only set the retry count if a dead letter sink is specified.
		deliverySpec.Retry = deliverySpecDefaults.Retry(ns)
	}
	// Besides this, the eventing webhook will add the usual defaults.
	return deliverySpec
}
