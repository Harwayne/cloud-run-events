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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var (
	channelBrokerConditionReady = apis.Condition{
		Type:   ChannelBrokerConditionReady,
		Status: corev1.ConditionTrue,
	}

	channelBrokerConditionIngress = apis.Condition{
		Type:   ChannelBrokerConditionIngress,
		Status: corev1.ConditionTrue,
	}

	channelBrokerConditionIngressFalse = apis.Condition{
		Type:   ChannelBrokerConditionIngress,
		Status: corev1.ConditionFalse,
	}

	channelBrokerConditionFanout = apis.Condition{
		Type:   ChannelBrokerConditionFanout,
		Status: corev1.ConditionTrue,
	}

	channelBrokerConditionRetry = apis.Condition{
		Type:   ChannelBrokerConditionRetry,
		Status: corev1.ConditionTrue,
	}
)

func TestChannelBrokerGetCondition(t *testing.T) {
	tests := []struct {
		name      string
		ts        *ChannelBrokerStatus
		condQuery apis.ConditionType
		want      *apis.Condition
	}{{
		name: "single condition",
		ts: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{
					brokerCellConditionReady,
				},
			},
		},
		condQuery: apis.ConditionReady,
		want:      &brokerCellConditionReady,
	}, {
		name: "multiple conditions",
		ts: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{
					brokerCellConditionIngress,
					brokerCellConditionFanout,
				},
			},
		},
		condQuery: ChannelBrokerConditionIngress,
		want:      &brokerCellConditionIngress,
	}, {
		name: "multiple conditions, condition false",
		ts: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{
					brokerCellConditionIngressFalse,
					brokerCellConditionFanout,
				},
			},
		},
		condQuery: ChannelBrokerConditionIngress,
		want:      &brokerCellConditionIngressFalse,
	}, {
		name: "unknown condition",
		ts: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{
					brokerCellConditionIngress,
				},
			},
		},
		condQuery: apis.ConditionType("foo"),
		want:      nil,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.ts.GetCondition(test.condQuery)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected condition (-want, +got) = %v", diff)
			}
		})
	}
}

func TestChannelBrokerInitializeConditions(t *testing.T) {
	tests := []struct {
		name string
		ts   *ChannelBrokerStatus
		want *ChannelBrokerStatus
	}{{
		name: "empty",
		ts:   &ChannelBrokerStatus{},
		want: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{{
					Type:   ChannelBrokerConditionFanout,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionIngress,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionReady,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionRetry,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionTargetsConfig,
					Status: corev1.ConditionUnknown,
				}},
			},
		},
	}, {
		name: "one false",
		ts: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{{
					Type:   ChannelBrokerConditionIngress,
					Status: corev1.ConditionFalse,
				}},
			},
		},
		want: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{{
					Type:   ChannelBrokerConditionFanout,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionIngress,
					Status: corev1.ConditionFalse,
				}, {
					Type:   ChannelBrokerConditionReady,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionRetry,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionTargetsConfig,
					Status: corev1.ConditionUnknown,
				}},
			},
		},
	}, {
		name: "one true",
		ts: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{{
					Type:   ChannelBrokerConditionIngress,
					Status: corev1.ConditionTrue,
				}},
			},
		},
		want: &ChannelBrokerStatus{
			Status: duckv1.Status{
				Conditions: []apis.Condition{{
					Type:   ChannelBrokerConditionFanout,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionIngress,
					Status: corev1.ConditionTrue,
				}, {
					Type:   ChannelBrokerConditionReady,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionRetry,
					Status: corev1.ConditionUnknown,
				}, {
					Type:   ChannelBrokerConditionTargetsConfig,
					Status: corev1.ConditionUnknown,
				}},
			},
		},
	}}

	ignoreAllButTypeAndStatus := cmpopts.IgnoreFields(
		apis.Condition{},
		"LastTransitionTime", "Message", "Reason", "Severity")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.ts.InitializeConditions()
			if diff := cmp.Diff(test.want, test.ts, ignoreAllButTypeAndStatus); diff != "" {
				t.Errorf("unexpected conditions (-want, +got) = %v", diff)
			}
		})
	}
}

func TestChannelBrokerConditionStatus(t *testing.T) {
	tests := []struct {
		name                string
		fanoutStatus        *appsv1.Deployment
		ingressStatus       *corev1.Endpoints
		retryStatus         *appsv1.Deployment
		targetsStatus       bool
		wantConditionStatus corev1.ConditionStatus
	}{{
		name:                "all happy",
		fanoutStatus:        TestHelper.AvailableDeployment(),
		ingressStatus:       TestHelper.AvailableEndpoints(),
		retryStatus:         TestHelper.AvailableDeployment(),
		targetsStatus:       true,
		wantConditionStatus: corev1.ConditionTrue,
	}, {
		name:                "fanout sad",
		fanoutStatus:        TestHelper.UnavailableDeployment(),
		ingressStatus:       TestHelper.AvailableEndpoints(),
		retryStatus:         TestHelper.AvailableDeployment(),
		targetsStatus:       true,
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "fanout unknown",
		fanoutStatus:        TestHelper.UnknownDeployment(),
		ingressStatus:       TestHelper.AvailableEndpoints(),
		retryStatus:         TestHelper.AvailableDeployment(),
		targetsStatus:       true,
		wantConditionStatus: corev1.ConditionUnknown,
	}, {
		name:                "ingress sad",
		fanoutStatus:        TestHelper.AvailableDeployment(),
		ingressStatus:       TestHelper.UnavailableEndpoints(),
		retryStatus:         TestHelper.AvailableDeployment(),
		targetsStatus:       true,
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "retry sad",
		fanoutStatus:        TestHelper.AvailableDeployment(),
		ingressStatus:       TestHelper.AvailableEndpoints(),
		retryStatus:         TestHelper.UnavailableDeployment(),
		targetsStatus:       true,
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "retry unknown",
		fanoutStatus:        TestHelper.AvailableDeployment(),
		ingressStatus:       TestHelper.AvailableEndpoints(),
		retryStatus:         TestHelper.UnknownDeployment(),
		targetsStatus:       true,
		wantConditionStatus: corev1.ConditionUnknown,
	}, {
		name:                "targets sad",
		fanoutStatus:        TestHelper.AvailableDeployment(),
		ingressStatus:       TestHelper.AvailableEndpoints(),
		retryStatus:         TestHelper.AvailableDeployment(),
		targetsStatus:       false,
		wantConditionStatus: corev1.ConditionFalse,
	}, {
		name:                "all sad",
		fanoutStatus:        TestHelper.UnavailableDeployment(),
		ingressStatus:       TestHelper.UnavailableEndpoints(),
		retryStatus:         TestHelper.UnavailableDeployment(),
		targetsStatus:       false,
		wantConditionStatus: corev1.ConditionFalse,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs := &ChannelBrokerStatus{}
			if test.fanoutStatus != nil {
				bs.PropagateFanoutAvailability(test.fanoutStatus)
			} else {
				bs.PropagateFanoutAvailability(&appsv1.Deployment{})
			}
			if test.ingressStatus != nil {
				bs.PropagateIngressAvailability(test.ingressStatus)
			} else {
				bs.PropagateIngressAvailability(&corev1.Endpoints{})
			}
			if test.retryStatus != nil {
				bs.PropagateRetryAvailability(test.retryStatus)
			} else {
				bs.PropagateRetryAvailability(&appsv1.Deployment{})
			}
			if test.targetsStatus {
				bs.MarkTargetsConfigReady()
			} else {
				bs.MarkTargetsConfigFailed("Unable to sync targets config", "induced failure")
			}
			got := bs.GetTopLevelCondition().Status
			if test.wantConditionStatus != got {
				t.Errorf("unexpected readiness: want %v, got %v", test.wantConditionStatus, got)
			}
			happy := bs.IsReady()
			switch test.wantConditionStatus {
			case corev1.ConditionTrue:
				if !happy {
					t.Error("expected happy true, got false")
				}
			case corev1.ConditionFalse, corev1.ConditionUnknown:
				if happy {
					t.Error("expected happy false, got true")
				}
			}
		})
	}
}
