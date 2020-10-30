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

	"knative.dev/pkg/injection"

	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	"github.com/google/knative-gcp/pkg/apis/configs/gcpauth"
	"github.com/google/knative-gcp/pkg/apis/messaging/v1beta1"
	brokerinformers "github.com/google/knative-gcp/pkg/client/injection/informers/broker/v1beta1/broker"
	triggerinformers "github.com/google/knative-gcp/pkg/client/injection/informers/broker/v1beta1/trigger"
	channelinformer "github.com/google/knative-gcp/pkg/client/injection/informers/messaging/v1beta1/channel"
	channelreconciler "github.com/google/knative-gcp/pkg/client/injection/reconciler/messaging/v1beta1/channel"
	"github.com/google/knative-gcp/pkg/reconciler"
	"github.com/google/knative-gcp/pkg/reconciler/identity"
	"github.com/google/knative-gcp/pkg/reconciler/identity/iam"
	serviceaccountinformers "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "Channels"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "cloud-run-events-channel-controller"
)

type Constructor injection.ControllerConstructor

// NewConstructor creates a constructor to make a Channel controller.
func NewConstructor(ipm iam.IAMPolicyManager, gcpas *gcpauth.StoreSingleton) Constructor {
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		return newController(ctx, cmw, ipm, gcpas.Store(ctx, cmw))
	}
}

func newController(
	ctx context.Context,
	cmw configmap.Watcher,
	ipm iam.IAMPolicyManager,
	gcpas *gcpauth.Store,
) *controller.Impl {
	channelInformer := channelinformer.Get(ctx)

	serviceAccountInformer := serviceaccountinformers.Get(ctx)
	brokerInformer := brokerinformers.Get(ctx)
	triggerInformer := triggerinformers.Get(ctx)

	r := &Reconciler{
		Base:          reconciler.NewBase(ctx, controllerAgentName, cmw),
		Identity:      identity.NewIdentity(ctx, ipm, gcpas),
		channelLister: channelInformer.Lister(),
		brokerLister:  brokerInformer.Lister(),
	}
	impl := channelreconciler.NewImpl(ctx, r)

	r.Logger.Info("Setting up event handlers")
	channelInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	channelGK := v1beta1.Kind("Channel")

	// The Channel itself is mapped to Brokers.
	brokerInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(channelGK),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	// The Subscriptions are mapped to Triggers.
	triggerInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(channelGK),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	serviceAccountInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(channelGK),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
