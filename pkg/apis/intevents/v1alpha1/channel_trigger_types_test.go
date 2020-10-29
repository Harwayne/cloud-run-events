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

	"knative.dev/pkg/apis"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestChannelTrigger_GetGroupVersionKind(t *testing.T) {
	want := schema.GroupVersionKind{
		Group:   "internal.events.cloud.google.com",
		Version: "v1alpha1",
		Kind:    "ChannelTrigger",
	}
	bc := ChannelTrigger{}
	got := bc.GetGroupVersionKind()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(GetGroupVersionKind (-want +got): %v", diff)
	}
}

func TestChannelTrigger_GetUntypedSpec(t *testing.T) {
	bc := ChannelTrigger{
		Spec: ChannelTriggerSpec{},
	}
	s := bc.GetUntypedSpec()
	if _, ok := s.(ChannelTriggerSpec); !ok {
		t.Errorf("untyped spec was not a BrokerSpec")
	}
}

func TestChannelTrigger_GetConditionSet(t *testing.T) {
	bc := &ChannelTrigger{}

	if got, want := bc.GetConditionSet().GetTopLevelConditionType(), apis.ConditionReady; got != want {
		t.Errorf("GetTopLevelCondition=%v, want=%v", got, want)
	}
}

func TestChannelTrigger_GetStatus(t *testing.T) {
	bc := &ChannelTrigger{
		Status: ChannelTriggerStatus{},
	}
	if got, want := bc.GetStatus(), &bc.Status.Status; got != want {
		t.Errorf("GetStatus=%v, want=%v", got, want)
	}
}
