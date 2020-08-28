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

// Code generated by injection-gen. DO NOT EDIT.

package fake

import (
	context "context"

	fake "github.com/google/knative-gcp/pkg/client/injection/informers/factory/fake"
	topic "github.com/google/knative-gcp/pkg/client/injection/informers/intevents/v1alpha1/topic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
)

var Get = topic.Get

func init() {
	injection.Fake.RegisterInformer(
		withInformer,
		metav1.GroupVersionResource{
			Group:    "internal.events.cloud.google.com",
			Version:  "v1alpha1",
			Resource: "topics",
		})
}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := fake.Get(ctx)
	inf := f.Internal().V1alpha1().Topics()
	return context.WithValue(ctx, topic.Key{}, inf), inf.Informer()
}
