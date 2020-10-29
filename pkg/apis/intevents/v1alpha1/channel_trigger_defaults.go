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
	"context"
)

// SetDefaults sets the default field values for a ChannelTrigger.
func (bc *ChannelTrigger) SetDefaults(ctx context.Context) {
	// Set defaults for the Spec.Components values.
	bc.Spec.SetDefaults(ctx)
}

// SetDefaults sets the default field values for a ChannelTriggerSpec.
func (bcs *ChannelTriggerSpec) SetDefaults(ctx context.Context) {
}
