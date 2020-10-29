/*
Copyright 2019 Google LLC

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

package resources

import (
	gcpv1beta1 "github.com/google/knative-gcp/pkg/apis/broker/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1beta1 "knative.dev/eventing/pkg/apis/duck/v1beta1"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	pkgduckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// PullSubscriptionArgs are the arguments needed to create a Channel Subscriber.
// Every field is required.
type TriggerArgs struct {
	Owner       kmeta.OwnerRefable
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	Subscriber  duckv1beta1.SubscriberSpec
}

// MakePullSubscription generates (but does not insert into K8s) the
// PullSubscription for Channels.
func MakeTrigger(args *TriggerArgs) *gcpv1beta1.Trigger {
	return &gcpv1beta1.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       args.Owner.GetObjectMeta().GetNamespace(),
			Name:            args.Name,
			Labels:          args.Labels,
			Annotations:     args.Annotations,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(args.Owner)},
		},
		Spec: v1beta1.TriggerSpec{
			Broker: args.Owner.GetObjectMeta().GetName(),
			// No filter, everything goes through.
			Filter: nil,
			Subscriber: pkgduckv1.Destination{
				URI: args.Subscriber.SubscriberURI,
			},
		},
		// TODO: Reply!
	}
}
