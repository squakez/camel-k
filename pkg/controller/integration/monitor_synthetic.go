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

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/trait"
)

// NewMonitorSyntheticAction is an action used to monitor synthetic Integrations.
func NewMonitorSyntheticAction() Action {
	return &monitorSyntheticAction{}
}

type monitorSyntheticAction struct {
	monitorAction
}

func (action *monitorSyntheticAction) Name() string {
	return "monitor-synthetic"
}

func (action *monitorSyntheticAction) Handle(ctx context.Context, integration *v1.Integration) (*v1.Integration, error) {
	environment, err := trait.NewSyntheticEnvironment(ctx, action.client, integration, nil)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Not an error: the resource from which we imported has been deleted, report in it status.
			// It may be a temporary situation, for example, if the deployment from which the Integration is imported
			// is being redeployed. For this reason we should keep the Integration instead of forcefully removing it.
			message := fmt.Sprintf(
				"import %s %s no longer available",
				integration.Annotations[v1.IntegrationImportedKindLabel],
				integration.Annotations[v1.IntegrationImportedNameLabel],
			)
			action.L.Info(message)
			integration.SetReadyConditionError(message)
			zero := int32(0)
			integration.Status.Phase = v1.IntegrationPhaseImportMissing
			integration.Status.Replicas = &zero
			return integration, nil
		}
		// report the error
		integration.Status.Phase = v1.IntegrationPhaseError
		integration.SetReadyCondition(corev1.ConditionFalse, v1.IntegrationConditionImportingKindAvailableReason, err.Error())
		return integration, err
	}

	return action.monitorPods(ctx, environment, integration)
}
