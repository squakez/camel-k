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

package integrationplatform

import (
	"context"
	"testing"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/platform"
	"github.com/apache/camel-k/v2/pkg/util/defaults"
	"github.com/apache/camel-k/v2/pkg/util/log"
	"github.com/apache/camel-k/v2/pkg/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
)

func TestCanHandlePhaseReadyOrError(t *testing.T) {
	ip := v1.IntegrationPlatform{}
	ip.Namespace = "ns"
	ip.Name = "ck"
	ip.Status.Phase = v1.IntegrationPlatformPhaseReady
	c, err := test.NewFakeClient(&ip)
	require.NoError(t, err)

	action := NewMonitorAction()
	action.InjectLogger(log.Log)
	action.InjectClient(c)

	answer := action.CanHandle(&ip)
	assert.True(t, answer)

	ip.Status.Phase = v1.IntegrationPlatformPhaseError
	answer = action.CanHandle(&ip)
	assert.True(t, answer)

	ip.Status.Phase = v1.IntegrationPlatformPhaseCreating
	answer = action.CanHandle(&ip)
	assert.False(t, answer)
}

func TestMonitorReady(t *testing.T) {
	ip := v1.IntegrationPlatform{}
	ip.Namespace = "ns"
	ip.Name = "ck"
	ip.Spec.Build.RuntimeVersion = "1.2.3"
	ip.Spec.Build.RuntimeProvider = v1.RuntimeProviderQuarkus
	ip.Status.Build.RuntimeVersion = "1.2.3"
	ip.Status.Build.RuntimeProvider = v1.RuntimeProviderQuarkus
	ip.Status.Build.Registry.Address = "1.2.3.4"
	ip.Status.Phase = v1.IntegrationPlatformPhaseReady
	c, err := test.NewFakeClient(&ip)
	require.NoError(t, err)

	action := NewMonitorAction()
	action.InjectLogger(log.Log)
	action.InjectClient(c)

	answer, err := action.Handle(context.TODO(), &ip)
	require.NoError(t, err)
	assert.NotNil(t, answer)

	assert.Equal(t, v1.IntegrationPlatformPhaseReady, answer.Status.Phase)
	assert.Equal(t, corev1.ConditionTrue, answer.Status.GetCondition(v1.IntegrationPlatformConditionTypeRegistryAvailable).Status)
}

func TestMonitorDrift(t *testing.T) {
	ip := v1.IntegrationPlatform{}
	ip.Namespace = "ns"
	ip.Name = "ck"
	ip.Spec.Build.RuntimeVersion = "3.2.1"
	ip.Spec.Build.RuntimeProvider = v1.RuntimeProviderQuarkus
	ip.Status.Build.RuntimeVersion = "1.2.3"
	ip.Status.Build.RuntimeProvider = v1.RuntimeProviderQuarkus
	ip.Status.Build.Registry.Address = "1.2.3.4"
	ip.Status.Phase = v1.IntegrationPlatformPhaseReady
	c, err := test.NewFakeClient(&ip)
	require.NoError(t, err)

	action := NewMonitorAction()
	action.InjectLogger(log.Log)
	action.InjectClient(c)

	answer, err := action.Handle(context.TODO(), &ip)
	require.NoError(t, err)
	assert.NotNil(t, answer)

	assert.Equal(t, v1.IntegrationPlatformPhaseNone, answer.Status.Phase)
}

func TestMonitorDriftDefault(t *testing.T) {
	ip := v1.IntegrationPlatform{}
	ip.Namespace = "ns"
	ip.Name = "ck"
	ip.Status.Build.RuntimeVersion = defaults.DefaultRuntimeVersion
	ip.Status.Build.RuntimeProvider = v1.RuntimeProviderQuarkus
	ip.Status.Build.Registry.Address = "1.2.3.4"
	ip.Status.Phase = v1.IntegrationPlatformPhaseReady
	c, err := test.NewFakeClient(&ip)
	require.NoError(t, err)

	action := NewMonitorAction()
	action.InjectLogger(log.Log)
	action.InjectClient(c)

	answer, err := action.Handle(context.TODO(), &ip)
	require.NoError(t, err)
	assert.NotNil(t, answer)

	assert.Equal(t, v1.IntegrationPlatformPhaseReady, answer.Status.Phase)
}

func TestMonitorMissingRegistryError(t *testing.T) {
	ip := v1.IntegrationPlatform{}
	ip.Namespace = "ns"
	ip.Name = "ck"
	ip.Spec.Build.RuntimeVersion = "1.2.3"
	ip.Spec.Build.RuntimeProvider = v1.RuntimeProviderQuarkus
	ip.Status.Build.RuntimeVersion = "1.2.3"
	ip.Status.Build.RuntimeProvider = v1.RuntimeProviderQuarkus
	c, err := test.NewFakeClient(&ip)
	require.NoError(t, err)

	err = platform.ConfigureDefaults(context.TODO(), c, &ip, false)
	require.NoError(t, err)

	action := NewMonitorAction()
	action.InjectLogger(log.Log)
	action.InjectClient(c)

	answer, err := action.Handle(context.TODO(), &ip)
	require.NoError(t, err)
	assert.NotNil(t, answer)

	assert.Equal(t, v1.IntegrationPlatformPhaseError, answer.Status.Phase)
	assert.Equal(t, corev1.ConditionFalse, answer.Status.GetCondition(v1.IntegrationPlatformConditionTypeRegistryAvailable).Status)
	assert.Equal(t, v1.IntegrationPlatformConditionTypeRegistryAvailableReason, answer.Status.GetCondition(v1.IntegrationPlatformConditionTypeRegistryAvailable).Reason)
	assert.Equal(t, "registry address not available, you need to set one", answer.Status.GetCondition(v1.IntegrationPlatformConditionTypeRegistryAvailable).Message)
}
