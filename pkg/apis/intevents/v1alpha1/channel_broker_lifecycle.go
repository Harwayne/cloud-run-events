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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
)

const (
	// ChannelBrokerConditionReady has status true when all subconditions below
	// have been set to True.
	ChannelBrokerConditionReady apis.ConditionType = apis.ConditionReady

	// ChannelBrokerConditionIngress reports the availability of the
	// ChannelBroker's ingress service.
	ChannelBrokerConditionIngress apis.ConditionType = "IngressReady"

	// ChannelBrokerConditionFanout reports the readiness of the ChannelBroker's
	// fanout service.
	ChannelBrokerConditionFanout apis.ConditionType = "FanoutReady"

	// ChannelBrokerConditionRetry reports the readiness of the ChannelBroker's retry
	// service.
	ChannelBrokerConditionRetry apis.ConditionType = "RetryReady"

	// ChannelBrokerConditionTargetsConfig reports the readiness of the
	// ChannelBroker's targets configmap.
	ChannelBrokerConditionTargetsConfig apis.ConditionType = "TargetsConfigReady"
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (bs *ChannelBrokerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return brokerCellCondSet.Manage(bs).GetCondition(t)
}

// GetTopLevelCondition returns the top level Condition.
func (bs *ChannelBrokerStatus) GetTopLevelCondition() *apis.Condition {
	return brokerCellCondSet.Manage(bs).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (bs *ChannelBrokerStatus) IsReady() bool {
	return brokerCellCondSet.Manage(bs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (bs *ChannelBrokerStatus) InitializeConditions() {
	brokerCellCondSet.Manage(bs).InitializeConditions()
}

// PropagateIngressAvailability uses the availability of the provided Endpoints
// to determine if ChannelBrokerConditionIngress should be marked as true or
// false.
func (bs *ChannelBrokerStatus) PropagateIngressAvailability(ep *corev1.Endpoints) {
	if duck.EndpointsAreAvailable(ep) {
		brokerCellCondSet.Manage(bs).MarkTrue(ChannelBrokerConditionIngress)
	} else {
		brokerCellCondSet.Manage(bs).MarkFalse(ChannelBrokerConditionIngress, "EndpointsUnavailable", "Endpoints %q is unavailable.", ep.Name)
	}
}

func (bs *ChannelBrokerStatus) MarkIngressFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(ChannelBrokerConditionIngress, reason, format, args...)
}

// PropagateFanoutAvailability uses the availability of the provided Deployment
// to determine if ChannelBrokerConditionFanout should be marked as true or
// false.
func (bs *ChannelBrokerStatus) PropagateFanoutAvailability(d *appsv1.Deployment) {
	deploymentAvailableFound := false
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			deploymentAvailableFound = true
			if cond.Status == corev1.ConditionTrue {
				brokerCellCondSet.Manage(bs).MarkTrue(ChannelBrokerConditionFanout)
			} else if cond.Status == corev1.ConditionFalse {
				brokerCellCondSet.Manage(bs).MarkFalse(ChannelBrokerConditionFanout, cond.Reason, cond.Message)
			} else if cond.Status == corev1.ConditionUnknown {
				brokerCellCondSet.Manage(bs).MarkUnknown(ChannelBrokerConditionFanout, cond.Reason, cond.Message)
			}
		}
	}
	if !deploymentAvailableFound {
		brokerCellCondSet.Manage(bs).MarkUnknown(ChannelBrokerConditionFanout, "DeploymentUnavailable", "Deployment %q is unavailable.", d.Name)
	}
}

func (bs *ChannelBrokerStatus) MarkFanoutFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(ChannelBrokerConditionFanout, reason, format, args...)
}

func (bs *ChannelBrokerStatus) MarkFanoutUnknown(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkUnknown(ChannelBrokerConditionFanout, reason, format, args...)
}

// PropagateRetryAvailability uses the availability of the provided Deployment
// to determine if ChannelBrokerConditionRetry should be marked as true or
// unknown.
func (bs *ChannelBrokerStatus) PropagateRetryAvailability(d *appsv1.Deployment) {
	deploymentAvailableFound := false
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			deploymentAvailableFound = true
			if cond.Status == corev1.ConditionTrue {
				brokerCellCondSet.Manage(bs).MarkTrue(ChannelBrokerConditionRetry)
			} else if cond.Status == corev1.ConditionFalse {
				brokerCellCondSet.Manage(bs).MarkFalse(ChannelBrokerConditionRetry, cond.Reason, cond.Message)
			} else if cond.Status == corev1.ConditionUnknown {
				brokerCellCondSet.Manage(bs).MarkUnknown(ChannelBrokerConditionRetry, cond.Reason, cond.Message)
			}
		}
	}
	if !deploymentAvailableFound {
		brokerCellCondSet.Manage(bs).MarkUnknown(ChannelBrokerConditionRetry, "DeploymentUnavailable", "Deployment %q is unavailable.", d.Name)
	}
}

func (bs *ChannelBrokerStatus) MarkRetryFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(ChannelBrokerConditionRetry, reason, format, args...)
}

func (bs *ChannelBrokerStatus) MarkRetryUnknown(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkUnknown(ChannelBrokerConditionRetry, reason, format, args...)
}

func (bs *ChannelBrokerStatus) MarkTargetsConfigReady() {
	brokerCellCondSet.Manage(bs).MarkTrue(ChannelBrokerConditionTargetsConfig)
}

func (bs *ChannelBrokerStatus) MarkTargetsConfigFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(ChannelBrokerConditionTargetsConfig, reason, format, args...)
}
