/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"context"
	"fmt"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/trait"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

// NewImportAction creates a new import action.
func NewImportAction() Action {
	return &importAction{}
}

type importAction struct {
	baseAction
}

// Name returns a common name of the action.
func (action *importAction) Name() string {
	return "initialize"
}

// CanHandle tells whether this action can handle the integration.
func (action *importAction) CanHandle(integration *v1.Integration) bool {
	return integration.Status.Phase == v1.IntegrationPhaseImporting
}

// Handle handles the integrations.
func (action *importAction) Handle(ctx context.Context, it *v1.Integration) (*v1.Integration, error) {
	action.L.Info("Importing from existing deployment")
	// Reverse trait logic (from environment to Integration)
	newIt, err := trait.Reverse(ctx, action.client, it)
	if err != nil {
		newIt.Status.Phase = v1.IntegrationPhaseError
		newIt.SetReadyCondition(corev1.ConditionFalse,
			v1.IntegrationConditionInitializationFailedReason, err.Error())
		return newIt, err
	}
	// Patch the integration with the traits configuration got from the reverse func
	patch := ctrl.MergeFrom(it)
	err = action.client.Patch(ctx, newIt, patch)
	if err != nil {
		return newIt, err
	}
	newIt.Status.Phase = v1.IntegrationPhaseInitialization
	newIt.Status.SetCondition(
		v1.IntegrationConditionImported,
		corev1.ConditionTrue,
		v1.IntegrationConditionImportReason,
		fmt.Sprintf("Imported from deployment %s", newIt.Annotations["camel.apache.org/imported-by"]),
	)

	return newIt, nil
}
