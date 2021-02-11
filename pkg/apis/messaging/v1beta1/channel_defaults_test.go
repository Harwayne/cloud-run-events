/*
Copyright 2019 The Knative Authors

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
	"testing"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/google/knative-gcp/pkg/apis/configs/brokerdelivery"

	eventingduckv1beta1 "knative.dev/eventing/pkg/apis/duck/v1beta1"

	gcpauthtesthelper "github.com/google/knative-gcp/pkg/apis/configs/gcpauth/testhelper"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	clusterDefaultedBackoffDelay        = "PT1S"
	clusterDefaultedBackoffPolicy       = eventingduckv1beta1.BackoffPolicyExponential
	clusterDefaultedRetry         int32 = 6
	nsDefaultedBackoffDelay             = "PT2S"
	nsDefaultedBackoffPolicy            = eventingduckv1beta1.BackoffPolicyLinear
	nsDefaultedRetry              int32 = 10
	customRetry                   int32 = 5

	defaultConfig = &brokerdelivery.Config{
		BrokerDeliverySpecDefaults: &brokerdelivery.Defaults{
			// NamespaceDefaultsConfig are the default Broker Configs for each namespace.
			// Namespace is the key, the value is the KReference to the config.
			NamespaceDefaults: map[string]brokerdelivery.ScopedDefaults{
				"mynamespace": {
					DeliverySpec: &eventingduckv1beta1.DeliverySpec{
						BackoffDelay:  &nsDefaultedBackoffDelay,
						BackoffPolicy: &nsDefaultedBackoffPolicy,
						DeadLetterSink: &duckv1.Destination{
							URI: &apis.URL{
								Scheme: "pubsub",
								Host:   "ns-default-dead-letter-topic-id",
							},
						},
						Retry: &nsDefaultedRetry,
					},
				},
				"mynamespace2": {
					DeliverySpec: &eventingduckv1beta1.DeliverySpec{
						DeadLetterSink: &duckv1.Destination{
							URI: &apis.URL{
								Scheme: "pubsub",
								Host:   "ns-default-dead-letter-topic-id",
							},
						},
						Retry: &nsDefaultedRetry,
					},
				},
				"mynamespace3": {
					DeliverySpec: &eventingduckv1beta1.DeliverySpec{
						BackoffDelay:  &nsDefaultedBackoffDelay,
						BackoffPolicy: &nsDefaultedBackoffPolicy,
					},
				},
				"mynamespace4": {
					DeliverySpec: &eventingduckv1beta1.DeliverySpec{
						Retry: &nsDefaultedRetry,
					},
				},
			},
			ClusterDefaults: brokerdelivery.ScopedDefaults{
				DeliverySpec: &eventingduckv1beta1.DeliverySpec{
					BackoffDelay:  &clusterDefaultedBackoffDelay,
					BackoffPolicy: &clusterDefaultedBackoffPolicy,
					DeadLetterSink: &duckv1.Destination{
						URI: &apis.URL{
							Scheme: "pubsub",
							Host:   "cluster-default-dead-letter-topic-id",
						},
					},
					Retry: &clusterDefaultedRetry,
				},
			},
		},
	}
)

func TestChannelDefaults(t *testing.T) {
	for n, tc := range map[string]struct {
		in   Channel
		want Channel
	}{
		"no subscribable annotation": {
			in: Channel{},
			want: Channel{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						"messaging.knative.dev/subscribable": "v1beta1",
					},
				},
				Spec: ChannelSpec{},
			},
		},
		"with a subscribable annotation": {
			in: Channel{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						"messaging.knative.dev/subscribable": "v1beta1",
					},
				},
			},
			want: Channel{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						"messaging.knative.dev/subscribable": "v1beta1",
					},
				},
				Spec: ChannelSpec{},
			},
		},
		"DeliverySpec is defaulted": {
			in: Channel{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "mynamespace3",
					Annotations: map[string]string{
						"messaging.knative.dev/subscribable": "v1beta1",
					},
				},
				Spec: ChannelSpec{
					SubscribableSpec: &eventingduckv1beta1.SubscribableSpec{
						Subscribers: []eventingduckv1beta1.SubscriberSpec{
							{},
						},
					},
				},
			},
			want: Channel{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "mynamespace3",
					Annotations: map[string]string{
						"messaging.knative.dev/subscribable": "v1beta1",
					},
				},
				Spec: ChannelSpec{
					SubscribableSpec: &eventingduckv1beta1.SubscribableSpec{
						Subscribers: []eventingduckv1beta1.SubscriberSpec{
							{
								Delivery: &eventingduckv1beta1.DeliverySpec{
									BackoffDelay:  &nsDefaultedBackoffDelay,
									BackoffPolicy: &nsDefaultedBackoffPolicy,
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(n, func(t *testing.T) {
			ctx := gcpauthtesthelper.ContextWithDefaults()
			ctx = brokerdelivery.ToContext(ctx, defaultConfig)
			got := tc.in
			got.SetDefaults(ctx)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("failed to get expected (-want, +got) = %v", diff)
			}
		})
	}
}
