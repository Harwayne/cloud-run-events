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
	// ChannelTriggerConditionReady has status true when all subconditions below
	// have been set to True.
	ChannelTriggerConditionReady apis.ConditionType = apis.ConditionReady

	// ChannelTriggerConditionIngress reports the availability of the
	// ChannelTrigger's ingress service.
	ChannelTriggerConditionIngress apis.ConditionType = "IngressReady"

	// ChannelTriggerConditionFanout reports the readiness of the ChannelTrigger's
	// fanout service.
	ChannelTriggerConditionFanout apis.ConditionType = "FanoutReady"

	// ChannelTriggerConditionRetry reports the readiness of the ChannelTrigger's retry
	// service.
	ChannelTriggerConditionRetry apis.ConditionType = "RetryReady"

	// ChannelTriggerConditionTargetsConfig reports the readiness of the
	// ChannelTrigger's targets configmap.
	ChannelTriggerConditionTargetsConfig apis.ConditionType = "TargetsConfigReady"
)

// GetCondition returns the condition currently associated with the given type, or nil.
func (bs *ChannelTriggerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return brokerCellCondSet.Manage(bs).GetCondition(t)
}

// GetTopLevelCondition returns the top level Condition.
func (bs *ChannelTriggerStatus) GetTopLevelCondition() *apis.Condition {
	return brokerCellCondSet.Manage(bs).GetTopLevelCondition()
}

// IsReady returns true if the resource is ready overall.
func (bs *ChannelTriggerStatus) IsReady() bool {
	return brokerCellCondSet.Manage(bs).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (bs *ChannelTriggerStatus) InitializeConditions() {
	brokerCellCondSet.Manage(bs).InitializeConditions()
}

// PropagateIngressAvailability uses the availability of the provided Endpoints
// to determine if ChannelTriggerConditionIngress should be marked as true or
// false.
func (bs *ChannelTriggerStatus) PropagateIngressAvailability(ep *corev1.Endpoints) {
	if duck.EndpointsAreAvailable(ep) {
		brokerCellCondSet.Manage(bs).MarkTrue(ChannelTriggerConditionIngress)
	} else {
		brokerCellCondSet.Manage(bs).MarkFalse(ChannelTriggerConditionIngress, "EndpointsUnavailable", "Endpoints %q is unavailable.", ep.Name)
	}
}

func (bs *ChannelTriggerStatus) MarkIngressFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(ChannelTriggerConditionIngress, reason, format, args...)
}

// PropagateFanoutAvailability uses the availability of the provided Deployment
// to determine if ChannelTriggerConditionFanout should be marked as true or
// false.
func (bs *ChannelTriggerStatus) PropagateFanoutAvailability(d *appsv1.Deployment) {
	deploymentAvailableFound := false
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			deploymentAvailableFound = true
			if cond.Status == corev1.ConditionTrue {
				brokerCellCondSet.Manage(bs).MarkTrue(ChannelTriggerConditionFanout)
			} else if cond.Status == corev1.ConditionFalse {
				brokerCellCondSet.Manage(bs).MarkFalse(ChannelTriggerConditionFanout, cond.Reason, cond.Message)
			} else if cond.Status == corev1.ConditionUnknown {
				brokerCellCondSet.Manage(bs).MarkUnknown(ChannelTriggerConditionFanout, cond.Reason, cond.Message)
			}
		}
	}
	if !deploymentAvailableFound {
		brokerCellCondSet.Manage(bs).MarkUnknown(ChannelTriggerConditionFanout, "DeploymentUnavailable", "Deployment %q is unavailable.", d.Name)
	}
}

func (bs *ChannelTriggerStatus) MarkFanoutFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(ChannelTriggerConditionFanout, reason, format, args...)
}

func (bs *ChannelTriggerStatus) MarkFanoutUnknown(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkUnknown(ChannelTriggerConditionFanout, reason, format, args...)
}

// PropagateRetryAvailability uses the availability of the provided Deployment
// to determine if ChannelTriggerConditionRetry should be marked as true or
// unknown.
func (bs *ChannelTriggerStatus) PropagateRetryAvailability(d *appsv1.Deployment) {
	deploymentAvailableFound := false
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			deploymentAvailableFound = true
			if cond.Status == corev1.ConditionTrue {
				brokerCellCondSet.Manage(bs).MarkTrue(ChannelTriggerConditionRetry)
			} else if cond.Status == corev1.ConditionFalse {
				brokerCellCondSet.Manage(bs).MarkFalse(ChannelTriggerConditionRetry, cond.Reason, cond.Message)
			} else if cond.Status == corev1.ConditionUnknown {
				brokerCellCondSet.Manage(bs).MarkUnknown(ChannelTriggerConditionRetry, cond.Reason, cond.Message)
			}
		}
	}
	if !deploymentAvailableFound {
		brokerCellCondSet.Manage(bs).MarkUnknown(ChannelTriggerConditionRetry, "DeploymentUnavailable", "Deployment %q is unavailable.", d.Name)
	}
}

func (bs *ChannelTriggerStatus) MarkRetryFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(ChannelTriggerConditionRetry, reason, format, args...)
}

func (bs *ChannelTriggerStatus) MarkRetryUnknown(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkUnknown(ChannelTriggerConditionRetry, reason, format, args...)
}

func (bs *ChannelTriggerStatus) MarkTargetsConfigReady() {
	brokerCellCondSet.Manage(bs).MarkTrue(ChannelTriggerConditionTargetsConfig)
}

func (bs *ChannelTriggerStatus) MarkTargetsConfigFailed(reason, format string, args ...interface{}) {
	brokerCellCondSet.Manage(bs).MarkFalse(ChannelTriggerConditionTargetsConfig, reason, format, args...)
}
