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
	brokerv1beta1 "github.com/google/knative-gcp/pkg/apis/broker/v1beta1"
	gcpv1beta1 "github.com/google/knative-gcp/pkg/apis/broker/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "knative.dev/eventing/pkg/apis/eventing/v1"

	"knative.dev/pkg/kmeta"
)

// TopicArgs are the arguments needed to create a Channel Topic.
// Every field is required.
type BrokerArgs struct {
	Owner       kmeta.OwnerRefable
	Name        string
	Labels      map[string]string
	Annotations map[string]string
}

// MakeInvoker generates (but does not insert into K8s) the Topic for Channels.
func MakeBroker(args *BrokerArgs) *gcpv1beta1.Broker {
	if args.Annotations == nil {
		args.Annotations = make(map[string]string)
	}
	// It must be a Google Broker.
	args.Annotations[v1.BrokerClassAnnotationKey] = brokerv1beta1.BrokerClass

	return &gcpv1beta1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       args.Owner.GetObjectMeta().GetNamespace(),
			Name:            args.Name,
			Labels:          args.Labels,
			Annotations:     args.Annotations,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(args.Owner)},
		},
	}
}
