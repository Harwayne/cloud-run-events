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

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	nethttp "net/http"
	"sync"

	"go.uber.org/zap"
	"knative.dev/eventing/pkg/tracing"
	tracingconfig "knative.dev/pkg/tracing/config"

	"go.opencensus.io/plugin/ochttp"
	"knative.dev/pkg/tracing/propagation/tracecontextb3"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"

	"go.opencensus.io/trace"

	"github.com/kelseyhightower/envconfig"

	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/knative-gcp/pkg/kncloudevents"
	"github.com/google/knative-gcp/test/e2e/lib"
)

type envConfig struct {
	BrokerURL string `envconfig:"BROKER_URL" required:"true"`
}

type Receiver struct {
	client    cloudevents.Client
	errsCount int
	mux       sync.Mutex
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		panic(fmt.Sprintf("Failed to process env var: %s", err))
	}
	client, err := kncloudevents.NewDefaultClient(env.BrokerURL)
	if err != nil {
		panic(err)
	}
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	tracing.SetupStaticPublishing(logger.Sugar(), "localhost", &tracingconfig.Config{
		Backend:    tracingconfig.Stackdriver,
		SampleRate: 1.0,
	})
	r := &Receiver{
		client:    client,
		errsCount: 0,
	}
	h := ochttp.Handler{
		Handler:     r,
		Propagation: tracecontextb3.B3Egress,
	}
	if err := nethttp.ListenAndServe(":8080", &h); err != nil {
		log.Fatal(err)
	}
}

func (r *Receiver) ServeHTTP(response nethttp.ResponseWriter, request *nethttp.Request) {
	event, err := toEvent(request.Context(), request)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		log.Printf("Error getting the event from the request: %v", err)
		return
	}
	span := trace.FromContext(request.Context())
	log.Printf("Incoming span: %+v", span)
	// Check if the received event is the dummy event sent by sender pod.
	// If it is, send back a response CloudEvent.
	if event.ID() == lib.E2EDummyEventID {
		oobEvent := cloudevents.NewEvent(cloudevents.VersionV1)
		oobEvent.SetID(lib.E2EDummyRespEventID)
		oobEvent.SetType(lib.E2EDummyRespEventType)
		oobEvent.SetSource(lib.E2EDummyRespEventSource)
		oobEvent.SetData(cloudevents.ApplicationJSON, `{"source": "oob_sender!"}`)
		err := r.sendEvent(request.Context(), oobEvent)
		if err != nil {
			// Ksvc seems to auto retry 5xx. So use 4xx for predictability.
			response.WriteHeader(http.StatusFailedDependency)
			return
		}
		response.WriteHeader(http.StatusAccepted)
	} else {
		response.WriteHeader(http.StatusForbidden)
	}
}

func (r *Receiver) sendEvent(ctx context.Context, event cloudevents.Event) error {
	ctx = cloudevents.WithEncodingBinary(ctx)
	span := trace.FromContext(ctx)
	log.Printf("Just before going out span: %+v", span)
	result := r.client.Send(ctx, event)
	if cloudevents.IsACK(result) {
		return nil
	}
	return fmt.Errorf("event send did not ACK: %w", result)
}

// toEvent converts an http request to an event.
func toEvent(ctx context.Context, request *nethttp.Request) (*cloudevents.Event, error) {
	message := cehttp.NewMessageFromHttpRequest(request)
	defer func() {
		if err := message.Finish(nil); err != nil {
			log.Println("Failed to close message")
		}
	}()
	// If encoding is unknown, the message is not an event.
	if message.ReadEncoding() == binding.EncodingUnknown {
		return nil, errors.New("unknown encoding. Not a cloud event? ")
	}
	event, err := binding.ToEvent(request.Context(), message, transformer.AddTimeNow)
	if err != nil {
		return nil, err
	}
	return event, nil
}
