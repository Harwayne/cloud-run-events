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

package channel

import (
	"context"
	"fmt"

	gcplisters "github.com/google/knative-gcp/pkg/client/listers/broker/v1beta1"

	gcpv1beta1 "github.com/google/knative-gcp/pkg/apis/broker/v1beta1"

	v1 "knative.dev/eventing/pkg/apis/eventing/v1"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	eventingduckv1beta1 "knative.dev/eventing/pkg/apis/duck/v1beta1"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	"github.com/google/knative-gcp/pkg/apis/duck"
	"github.com/google/knative-gcp/pkg/apis/messaging/v1beta1"
	channelreconciler "github.com/google/knative-gcp/pkg/client/injection/reconciler/messaging/v1beta1/channel"
	inteventslisters "github.com/google/knative-gcp/pkg/client/listers/intevents/v1beta1"
	listers "github.com/google/knative-gcp/pkg/client/listers/messaging/v1beta1"
	"github.com/google/knative-gcp/pkg/reconciler"
	"github.com/google/knative-gcp/pkg/reconciler/identity"
	"github.com/google/knative-gcp/pkg/reconciler/messaging/channel/resources"
)

const (
	resourceGroup = "channels.messaging.cloud.google.com"

	reconciledSuccessReason                 = "ChannelReconciled"
	reconciledTopicFailedReason             = "TopicReconcileFailed"
	deleteWorkloadIdentityFailed            = "WorkloadIdentityDeleteFailed"
	reconciledSubscribersFailedReason       = "SubscribersReconcileFailed"
	reconciledSubscribersStatusFailedReason = "SubscribersStatusReconcileFailed"
	workloadIdentityFailed                  = "WorkloadIdentityReconcileFailed"
)

// Reconciler implements controller.Reconciler for Channel resources.
type Reconciler struct {
	*reconciler.Base
	// identity reconciler for reconciling workload identity.
	*identity.Identity
	// listers index properties about resources
	channelLister listers.ChannelLister
	topicLister   inteventslisters.TopicLister
	brokerLister  gcplisters.BrokerLister
}

// Check that our Reconciler implements Interface.
var _ channelreconciler.Interface = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, channel *v1beta1.Channel) pkgreconciler.Event {
	ctx = logging.WithLogger(ctx, r.Logger.With(zap.Any("channel", channel)))

	channel.Status.InitializeConditions()
	channel.Status.ObservedGeneration = channel.Generation

	// If ServiceAccountName is provided, reconcile workload identity.
	if channel.Spec.ServiceAccountName != "" {
		if _, err := r.Identity.ReconcileWorkloadIdentity(ctx, channel.Spec.Project, channel); err != nil {
			return pkgreconciler.NewEvent(corev1.EventTypeWarning, workloadIdentityFailed, "Failed to reconcile Channel workload identity: %s", err.Error())
		}
	}

	// 1. Create the Broker.
	broker, err := r.reconcileBroker(ctx, channel)
	if err != nil {
		channel.Status.MarkBrokerFailed("BrokerReconcileFailed", "Failed to reconcile Broker: %s", err.Error())
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciledTopicFailedReason, "Reconcile Broker failed with: %s", err.Error())
	}
	channel.Status.PropagateBrokerStatus(&broker.Status)

	// 2. Sync all subscriptions.
	//   a. create all Triggers that are in spec and not in status.
	//   b. delete all Triggers that are in status but not in spec.
	if err := r.syncSubscribers(ctx, channel); err != nil {
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciledSubscribersFailedReason, "Reconcile Subscribers failed with: %s", err.Error())
	}

	// 3. Sync all subscriptions statuses.
	if err := r.syncSubscribersStatus(ctx, channel); err != nil {
		return pkgreconciler.NewEvent(corev1.EventTypeWarning, reconciledSubscribersStatusFailedReason, "Reconcile Subscribers Status failed with: %s", err.Error())
	}

	return pkgreconciler.NewEvent(corev1.EventTypeNormal, reconciledSuccessReason, `Channel reconciled: "%s/%s"`, channel.Namespace, channel.Name)
}

func (r *Reconciler) syncSubscribers(ctx context.Context, channel *v1beta1.Channel) error {
	if channel.Status.SubscribableStatus.Subscribers == nil {
		channel.Status.SubscribableStatus.Subscribers = make([]eventingduckv1beta1.SubscriberStatus, 0)
	}

	subCreates := []eventingduckv1beta1.SubscriberSpec(nil)
	subUpdates := []eventingduckv1beta1.SubscriberSpec(nil)
	subDeletes := []eventingduckv1beta1.SubscriberStatus(nil)

	// Make a map of name to PullSubscription for lookup.
	pullsubs := make(map[string]gcpv1beta1.Trigger)
	if subs, err := r.getTriggers(ctx, channel); err != nil {
		logging.FromContext(ctx).Desugar().Error("Failed to list PullSubscriptions", zap.Error(err))
	} else {
		for _, s := range subs {
			pullsubs[s.Name] = s
		}
	}

	exists := make(map[types.UID]eventingduckv1beta1.SubscriberStatus)
	for _, s := range channel.Status.SubscribableStatus.Subscribers {
		exists[s.UID] = s
	}

	if channel.Spec.SubscribableSpec != nil {
		for _, want := range channel.Spec.SubscribableSpec.Subscribers {
			if got, ok := exists[want.UID]; !ok {
				// If it does not exist, then create it.
				subCreates = append(subCreates, want)
			} else {
				_, found := pullsubs[resources.GenerateTriggerName(want.UID)]
				// If did not find or the PS has updated generation, update it.
				if !found || got.ObservedGeneration != want.Generation {
					subUpdates = append(subUpdates, want)
				}
			}
			// Remove want from exists.
			delete(exists, want.UID)
		}
	}

	// Remaining exists will be deleted.
	for _, e := range exists {
		subDeletes = append(subDeletes, e)
	}

	clusterName := channel.GetAnnotations()[duck.ClusterNameAnnotation]
	for _, s := range subCreates {
		genName := resources.GenerateTriggerName(s.UID)

		t := resources.MakeTrigger(&resources.TriggerArgs{
			Owner:       channel,
			Name:        genName,
			Labels:      resources.GetPullSubscriptionLabels(controllerAgentName, channel.Name, genName, string(channel.UID)),
			Annotations: resources.GetPullSubscriptionAnnotations(channel.Name, clusterName),
			Subscriber:  s,
		})
		_, err := r.RunClientSet.EventingV1beta1().Triggers(channel.Namespace).Create(ctx, t, metav1.CreateOptions{})
		if apierrs.IsAlreadyExists(err) {
			// If the pullsub already exists and is owned by the current channel, mark it for update.
			if _, found := pullsubs[genName]; found {
				subUpdates = append(subUpdates, s)
			} else {
				r.Recorder.Eventf(channel, corev1.EventTypeWarning, "SubscriberNotOwned", "Subscriber %q is not owned by this channel", genName)
				return fmt.Errorf("channel %q does not own subscriber %q", channel.Name, genName)
			}
		} else if err != nil {
			r.Recorder.Eventf(channel, corev1.EventTypeWarning, "SubscriberCreateFailed", "Creating Subscriber %q failed", genName)
			return err
		}
		r.Recorder.Eventf(channel, corev1.EventTypeNormal, "SubscriberCreated", "Created Subscriber %q", genName)

		channel.Status.SubscribableStatus.Subscribers = append(channel.Status.SubscribableStatus.Subscribers, eventingduckv1beta1.SubscriberStatus{
			UID:                s.UID,
			ObservedGeneration: s.Generation,
		})
		return nil // Signal a re-reconcile.
	}
	for _, s := range subUpdates {
		genName := resources.GenerateTriggerName(s.UID)

		ps := resources.MakeTrigger(&resources.TriggerArgs{
			Owner:       channel,
			Name:        genName,
			Labels:      resources.GetPullSubscriptionLabels(controllerAgentName, channel.Name, genName, string(channel.UID)),
			Annotations: resources.GetPullSubscriptionAnnotations(channel.Name, clusterName),
			Subscriber:  s,
		})

		existingPs, found := pullsubs[genName]
		if !found {
			// PullSubscription does not exist, that's ok, create it now.
			ps, err := r.RunClientSet.EventingV1beta1().Triggers(channel.Namespace).Create(ctx, ps, metav1.CreateOptions{})
			if apierrs.IsAlreadyExists(err) {
				// If the pullsub is not owned by the current channel, this is an error.
				r.Recorder.Eventf(channel, corev1.EventTypeWarning, "SubscriberNotOwned", "Subscriber %q is not owned by this channel", genName)
				return fmt.Errorf("channel %q does not own subscriber %q", channel.Name, genName)
			} else if err != nil {
				r.Recorder.Eventf(channel, corev1.EventTypeWarning, "SubscriberCreateFailed", "Creating Subscriber %q failed", genName)
				return err
			}
			r.Recorder.Eventf(channel, corev1.EventTypeNormal, "SubscriberCreated", "Created Subscriber %q", ps.Name)
		} else if !equality.Semantic.DeepEqual(ps.Spec, existingPs.Spec) {
			// Don't modify the informers copy.
			desired := existingPs.DeepCopy()
			desired.Spec = ps.Spec
			ps, err := r.RunClientSet.EventingV1beta1().Triggers(channel.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
			if err != nil {
				r.Recorder.Eventf(channel, corev1.EventTypeWarning, "SubscriberUpdateFailed", "Updating Subscriber %q failed", genName)
				return err
			}
			r.Recorder.Eventf(channel, corev1.EventTypeNormal, "SubscriberUpdated", "Updated Subscriber %q", ps.Name)
		}
		for i, ss := range channel.Status.SubscribableStatus.Subscribers {
			if ss.UID == s.UID {
				channel.Status.SubscribableStatus.Subscribers[i].ObservedGeneration = s.Generation
				break
			}
		}
		return nil
	}
	for _, s := range subDeletes {
		genName := resources.GenerateTriggerName(s.UID)
		// TODO: we need to handle the case of a already deleted pull subscription. Perhaps move to ensure deleted method.
		if err := r.RunClientSet.InternalV1beta1().PullSubscriptions(channel.Namespace).Delete(ctx, genName, metav1.DeleteOptions{}); err != nil {
			logging.FromContext(ctx).Desugar().Error("unable to delete PullSubscription for Channel", zap.String("ps", genName), zap.String("channel", channel.Name), zap.Error(err))
			r.Recorder.Eventf(channel, corev1.EventTypeWarning, "SubscriberDeleteFailed", "Deleting Subscriber %q failed", genName)
			return err
		}
		r.Recorder.Eventf(channel, corev1.EventTypeNormal, "SubscriberDeleted", "Deleted Subscriber %q", genName)

		for i, ss := range channel.Status.SubscribableStatus.Subscribers {
			if ss.UID == s.UID {
				// Swap len-1 with i and then pop len-1 off the slice.
				channel.Status.SubscribableStatus.Subscribers[i] = channel.Status.SubscribableStatus.Subscribers[len(channel.Status.SubscribableStatus.Subscribers)-1]
				channel.Status.SubscribableStatus.Subscribers = channel.Status.SubscribableStatus.Subscribers[:len(channel.Status.SubscribableStatus.Subscribers)-1]
				break
			}
		}
		return nil // Signal a re-reconcile.
	}

	return nil
}

func (r *Reconciler) syncSubscribersStatus(ctx context.Context, channel *v1beta1.Channel) error {
	if channel.Status.SubscribableStatus.Subscribers == nil {
		channel.Status.SubscribableStatus.Subscribers = make([]eventingduckv1beta1.SubscriberStatus, 0)
	}

	// Make a map of subscriber name to PullSubscription for lookup.
	triggers := make(map[string]gcpv1beta1.Trigger)
	if subs, err := r.getTriggers(ctx, channel); err != nil {
		logging.FromContext(ctx).Desugar().Error("Failed to list Triggers", zap.Error(err))
	} else {
		for _, s := range subs {
			triggers[resources.ExtractUIDFromTriggerName(s.Name)] = s
		}
	}

	for i, ss := range channel.Status.SubscribableStatus.Subscribers {
		if ps, ok := triggers[string(ss.UID)]; ok {
			ready, msg := r.getTriggerStatus(&ps)
			channel.Status.SubscribableStatus.Subscribers[i].Ready = ready
			channel.Status.SubscribableStatus.Subscribers[i].Message = msg
		} else {
			logging.FromContext(ctx).Desugar().Error("Failed to find status for subscriber", zap.String("uid", string(ss.UID)))
		}
	}

	return nil
}

func (r *Reconciler) reconcileBroker(ctx context.Context, channel *v1beta1.Channel) (*gcpv1beta1.Broker, error) {
	clusterName := channel.GetAnnotations()[duck.ClusterNameAnnotation]
	name := fmt.Sprintf("chan-%s", channel.Name) // resources.GenerateBrokerName(channel)
	b := resources.MakeBroker(&resources.BrokerArgs{
		Owner:       channel,
		Name:        name,
		Labels:      resources.GetLabels(controllerAgentName, channel.Name, string(channel.UID)),
		Annotations: resources.GetTopicAnnotations(clusterName),
	})

	broker, err := r.getBroker(ctx, channel)
	if apierrs.IsNotFound(err) {
		broker, err = r.RunClientSet.EventingV1beta1().Brokers(channel.Namespace).Create(ctx, b, metav1.CreateOptions{})
		if err != nil {
			logging.FromContext(ctx).Desugar().Error("Failed to create Broker", zap.Error(err))
			r.Recorder.Eventf(channel, corev1.EventTypeWarning, "BrokerCreateFailed", "Failed to created Broker %q: %s", broker.Name, err.Error())
			return nil, err
		}
		r.Recorder.Eventf(channel, corev1.EventTypeNormal, "BrokerCreated", "Created Broker %q", broker.Name)
		return broker, nil
	} else if err != nil {
		logging.FromContext(ctx).Desugar().Error("Failed to get Broker", zap.Error(err))
		return nil, fmt.Errorf("failed to get Broker: %w", err)
	} else if !metav1.IsControlledBy(broker, channel) {
		channel.Status.MarkBrokerNotOwned("Broker %q is owned by another resource.", name)
		return nil, fmt.Errorf("Channel: %s does not own Broker: %s", channel.Name, name)
	} else if !equality.Semantic.DeepDerivative(b.Spec, broker.Spec) {
		// Don't modify the informers copy.
		desired := broker.DeepCopy()
		desired.Spec = b.Spec
		logging.FromContext(ctx).Desugar().Debug("Updating Broker", zap.Any("broker", desired))
		b, err = r.RunClientSet.EventingV1beta1().Brokers(channel.Namespace).Update(ctx, desired, metav1.UpdateOptions{})
		if err != nil {
			logging.FromContext(ctx).Desugar().Error("Failed to update Broker", zap.Any("broker", broker), zap.Error(err))
			return nil, fmt.Errorf("failed to update Broker: %w", err)
		}
		return b, nil
	} else if b.Annotations[v1.BrokerClassAnnotationKey] != gcpv1beta1.BrokerClass {
		panic("Bad brokerclass!")
	}

	if broker != nil {
		if broker.Status.Address.URL != nil {
			channel.Status.SetAddress(broker.Status.Address.URL)
		} else {
			channel.Status.SetAddress(nil)
		}
	}

	return broker, nil
}

func (r *Reconciler) getBroker(_ context.Context, channel *v1beta1.Channel) (*gcpv1beta1.Broker, error) {
	name := fmt.Sprintf("chan-%s", channel.Name) // resources.GenerateBrokerName(channel)
	broker, err := r.brokerLister.Brokers(channel.Namespace).Get(name)
	if err != nil {
		return nil, err
	}
	return broker, nil
}

func (r *Reconciler) getTriggers(ctx context.Context, channel *v1beta1.Channel) ([]gcpv1beta1.Trigger, error) {
	sl, err := r.RunClientSet.EventingV1beta1().Triggers(channel.Namespace).List(ctx, metav1.ListOptions{
		// Use GetLabelSelector to select all PullSubscriptions related to this channel.
		LabelSelector: resources.GetLabelSelector(controllerAgentName, channel.Name, string(channel.UID)).String(),
	})

	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Failed to list Triggers", zap.Error(err))
		return nil, err
	}
	triggers := make([]gcpv1beta1.Trigger, 0, len(sl.Items))
	for _, trigger := range sl.Items {
		if metav1.IsControlledBy(&trigger, channel) {
			triggers = append(triggers, trigger)
		}
	}
	return triggers, nil
}

func (r *Reconciler) getTriggerStatus(t *gcpv1beta1.Trigger) (corev1.ConditionStatus, string) {
	ready := corev1.ConditionTrue
	message := ""
	if !t.Status.IsReady() {
		ready = corev1.ConditionFalse
		message = fmt.Sprintf("Trigger %q is not ready", t.Name)
	}
	return ready, message
}

func (r *Reconciler) FinalizeKind(ctx context.Context, channel *v1beta1.Channel) pkgreconciler.Event {
	return nil
}
