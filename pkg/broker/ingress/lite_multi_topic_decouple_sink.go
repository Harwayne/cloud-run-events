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

package ingress

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"

	pubsublite2 "google.golang.org/genproto/googleapis/cloud/pubsublite/v1"

	pubsublite "cloud.google.com/go/pubsublite"
	pubsublitev1 "cloud.google.com/go/pubsublite/apiv1"

	"go.opencensus.io/trace"
	"go.uber.org/zap"

	cepubsub "github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	cev2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/google/knative-gcp/pkg/broker/config"
	"github.com/google/knative-gcp/pkg/logging"
)

// NewPubSubLiteMultiTopicDecoupleSink creates a new multiTopicDecoupleSink.
func NewPubSubLiteMultiTopicDecoupleSink(
	ctx context.Context,
	brokerConfig config.ReadonlyTargets,
	client *pubsublitev1.SubscriberClient,
	publishSettings pubsublite.PublishSettings) *pubSubLiteMultiTopicDecoupleSink {
	return &pubSubLiteMultiTopicDecoupleSink{
		pubsub:          client,
		publishSettings: publishSettings,
		brokerConfig:    brokerConfig,
		// TODO(#1118): remove Topic when broker config is removed
		topics: make(map[config.CellTenantKey]*pubsublitev1.PublisherClient),
		// TODO(#1804): remove this field when enabling the feature by default.
		enableEventFiltering: enableEventFilterFunc(),
	}
}

// multiTopicDecoupleSink implements DecoupleSink and routes events to pubsublite topics corresponding
// to the broker to which the events are sent.
type pubSubLiteMultiTopicDecoupleSink struct {
	// pubsublite talks to pubsublite.
	pubsub          *pubsublitev1.AdminClient
	publishSettings pubsublite.PublishSettings
	// map from brokers to topics
	topics    map[config.CellTenantKey]*pubSubLiteTopic
	topicsMut sync.RWMutex
	// brokerConfig holds configurations for all brokers. It's a view of a configmap populated by
	// the broker controller.
	brokerConfig config.ReadonlyTargets
	// TODO(#1804): remove this field when enabling the feature by default.
	enableEventFiltering bool
}

type pubSubLiteTopic struct {
	id     string
	close  func() error
	stream pubsublite2.PublisherService_PublishClient
}

// Send sends incoming event to its corresponding pubsublite topic based on which broker it belongs to.
func (m *pubSubLiteMultiTopicDecoupleSink) Send(ctx context.Context, broker *config.CellTenantKey, event cev2.Event) protocol.Result {
	topic, err := m.getTopicForBroker(ctx, broker)
	if err != nil {
		trace.FromContext(ctx).Annotate(
			[]trace.Attribute{
				trace.StringAttribute("error_message", err.Error()),
			},
			"unable to accept event",
		)
		return err
	}

	// Check to see if there are any triggers interested in this event. If not, no need to send this
	// to the decouple topic.
	// TODO(#1804): remove first check when enabling the feature by default.
	if m.enableEventFiltering && !m.hasTrigger(ctx, &event) {
		logging.FromContext(ctx).Debug("Filering target-less event at ingress", zap.String("Eventid", event.ID()))
		return nil
	}

	dt := extensions.FromSpanContext(trace.FromContext(ctx).SpanContext())
	msg := new(pubsub.Message)
	if err := cepubsub.WritePubSubMessage(ctx, binding.ToMessage(&event), msg, dt.WriteTransformer()); err != nil {
		return err
	}

	err = topic.stream.Send(&pubsublite2.PublishRequest{
		RequestType: &pubsublite2.PublishRequest_MessagePublishRequest{
			MessagePublishRequest: &pubsublite2.MessagePublishRequest{
				Messages: []*pubsublite2.PubSubMessage{
					{
						Key:  nil,
						Data: msg.Data,
						// TODO: Fix
						// Attributes: msg.Attributes,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = topic.stream.Recv()
	return err
}

// hasTrigger checks given event against all targets to see if it will pass any of their filters.
// If one is fouund, hasTrigger returns true.
func (m *pubSubLiteMultiTopicDecoupleSink) hasTrigger(ctx context.Context, event *cev2.Event) bool {
	hasTrigger := false
	m.brokerConfig.RangeAllTargets(func(target *config.Target) bool {
		if eventFilterFunc(ctx, target.FilterAttributes, event) {
			hasTrigger = true
			return false
		}

		return true
	})

	return hasTrigger
}

// getTopicForBroker finds the corresponding decouple topic for the broker from the mounted broker configmap volume.
func (m *pubSubLiteMultiTopicDecoupleSink) getTopicForBroker(ctx context.Context, broker *config.CellTenantKey) (*pubSubLiteTopic, error) {
	topicID, err := m.getTopicIDForBroker(ctx, broker)
	if err != nil {
		return nil, err
	}

	if topic, ok := m.getExistingTopic(broker); ok {
		// Check that the broker's topic ID hasn't changed.
		if topic.id == topicID {
			return topic, nil
		}
	}

	// Topic needs to be created or updated.
	return m.updateTopicForBroker(ctx, broker)
}

func (m *pubSubLiteMultiTopicDecoupleSink) updateTopicForBroker(ctx context.Context, broker *config.CellTenantKey) (*pubSubLiteTopic, error) {
	m.topicsMut.Lock()
	defer m.topicsMut.Unlock()
	// Fetch latest decouple topic ID under lock.
	topicID, err := m.getTopicIDForBroker(ctx, broker)
	if err != nil {
		return nil, err
	}

	if topic, ok := m.topics[*broker]; ok {
		if topic.id == topicID {
			// Topic already updated.
			return topic, nil
		}
		// Stop old topic.
		err := topic.close()
		if err != nil {
			logging.FromContext(ctx).Error("Error closing PubSubLite Topic", zap.Error(err), zap.String("topicID", topic.id))
		}
	}
	ps, err := pubsublitev1.NewPublisherClient(ctx)
	if err != nil {
		return nil, err
	}
	stream, err := ps.Publish(ctx)
	if err != nil {
		return nil, err
	}
	err = stream.Send(&pubsublite2.PublishRequest{})
	topic := &pubSubLiteTopic{
		id:     topicID,
		close:  ps.Close,
		stream: stream,
	}
	m.topics[*broker] = topic
	return topic, nil
}

func (m *pubSubLiteMultiTopicDecoupleSink) getTopicIDForBroker(ctx context.Context, broker *config.CellTenantKey) (string, error) {
	brokerConfig, ok := m.brokerConfig.GetCellTenantByKey(broker)
	if !ok {
		// There is an propagation delay between the controller reconciles the broker config and
		// the config being pushed to the configmap volume in the ingress pod. So sometimes we return
		// an error even if the request is valid.
		logging.FromContext(ctx).Warn("config is not found for")
		return "", fmt.Errorf("%q: %w", broker, ErrNotFound)
	}
	if brokerConfig.DecoupleQueue == nil || brokerConfig.DecoupleQueue.Topic == "" {
		logging.FromContext(ctx).Error("DecoupleQueue or topic missing for broker, this should NOT happen.", zap.Any("brokerConfig", brokerConfig))
		return "", fmt.Errorf("decouple queue of %q: %w", broker, ErrIncomplete)
	}
	if brokerConfig.DecoupleQueue.State != config.State_READY {
		logging.FromContext(ctx).Debug("decouple queue is not ready")
		return "", fmt.Errorf("%q: %w", broker, ErrNotReady)
	}
	return brokerConfig.DecoupleQueue.Topic, nil
}

func (m *pubSubLiteMultiTopicDecoupleSink) getExistingTopic(broker *config.CellTenantKey) (*pubSubLiteTopic, bool) {
	m.topicsMut.RLock()
	defer m.topicsMut.RUnlock()
	topic, ok := m.topics[*broker]
	return topic, ok
}
