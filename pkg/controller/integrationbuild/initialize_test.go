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

package integrationbuild

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/apis/camel/v1alpha1"
	"github.com/apache/camel-k/v2/pkg/internal"
	"github.com/apache/camel-k/v2/pkg/util/log"
)

func TestInitializeActionName(t *testing.T) {
	assert.Equal(t, "initialize", NewInitializeAction().Name())
}

func TestInitializeActionCanHandle(t *testing.T) {
	tests := []struct {
		phase    v1alpha1.IntegrationBuildPhase
		expected bool
	}{
		{v1alpha1.IntegrationBuildPhaseNone, true},
		{v1alpha1.IntegrationBuildPhaseBuilding, false},
		{v1alpha1.IntegrationBuildPhaseSucceeded, false},
		{v1alpha1.IntegrationBuildPhaseFailed, false},
	}

	a := NewInitializeAction()
	for _, tc := range tests {
		t.Run(string(tc.phase), func(t *testing.T) {
			assert.Equal(t, tc.expected, a.CanHandle(newTestIntegrationBuild(tc.phase)))
		})
	}
}

func TestInitializeActionHandle(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)
	c, err := internal.NewFakeClient(ib)
	require.NoError(t, err)

	a := NewInitializeAction()
	a.InjectLogger(log.Log)
	a.InjectClient(c)

	target, err := a.Handle(context.TODO(), ib)

	require.NoError(t, err)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseBuilding, target.Status.Phase)
	// The input must not be mutated
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseNone, ib.Status.Phase)

	var job batchv1.Job
	require.NoError(t, c.Get(context.TODO(), types.NamespacedName{Namespace: testNamespace, Name: testName}, &job))
}

func TestInitializeActionHandleCreateError(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)
	c, err := internal.NewFakeClient(ib)
	require.NoError(t, err)

	createErr := errors.New("boom")
	//nolint:forcetypeassert
	c.(*internal.FakeClient).Intercept(&interceptor.Funcs{
		Create: func(ctx context.Context, cl ctrl.WithWatch, obj ctrl.Object, opts ...ctrl.CreateOption) error {
			if _, ok := obj.(*batchv1.Job); ok {
				return createErr
			}

			return cl.Create(ctx, obj, opts...)
		},
	})

	a := NewInitializeAction()
	a.InjectLogger(log.Log)
	a.InjectClient(c)

	target, err := a.Handle(context.TODO(), ib)

	require.ErrorIs(t, err, createErr)
	assert.Contains(t, err.Error(), "cannot create IntegrationBuild Job")
	require.NotNil(t, target)
	// The build must move to the Error phase if the Job could not be created
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseError, target.Status.Phase)
}

func TestInitializeActionHandleJobAlreadyExists(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)
	existing := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace, Name: testName},
	}
	c, err := internal.NewFakeClient(ib, existing)
	require.NoError(t, err)

	a := NewInitializeAction()
	a.InjectLogger(log.Log)
	a.InjectClient(c)

	target, err := a.Handle(context.TODO(), ib)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot create IntegrationBuild Job")
	require.NotNil(t, target)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseError, target.Status.Phase)
}

const (
	missingBuildParamsMsg  = "imageBaseSource, imageDestination or imageDestinationVersion missing"
	missingExportParamsMsg = "runtimeProvider or gav missing"
)

var missingMandatoryParamsTests = []struct {
	name        string
	mutate      func(ib *v1alpha1.IntegrationBuild)
	expectedErr string
}{
	{
		"missing imageBaseSource",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.BuildTask.ImageBaseSource = "" },
		missingBuildParamsMsg,
	},
	{
		"missing imageDestination",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.BuildTask.ImageDestination = "" },
		missingBuildParamsMsg,
	},
	{
		"missing imageDestinationVersion",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.BuildTask.ImageDestinationVersion = "" },
		missingBuildParamsMsg,
	},
	{
		"missing buildTask",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.BuildTask = v1alpha1.MavenJIBBuilderTaskConf{} },
		missingBuildParamsMsg,
	},
	{
		"missing exportTask",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.ExportTask = nil },
		missingExportParamsMsg,
	},
	{
		"nil runtimeProvider",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.ExportTask.RuntimeProvider = nil },
		missingExportParamsMsg,
	},
	{
		"empty runtimeProvider",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.ExportTask.RuntimeProvider = ptr.To("") },
		missingExportParamsMsg,
	},
	{
		"nil gav",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.ExportTask.GAV = nil },
		missingExportParamsMsg,
	},
	{
		"empty gav",
		func(ib *v1alpha1.IntegrationBuild) { ib.Spec.ExportTask.GAV = ptr.To("") },
		missingExportParamsMsg,
	},
}

func TestInitializeActionHandleMissingMandatoryParams(t *testing.T) {
	for _, tc := range missingMandatoryParamsTests {
		t.Run(tc.name, func(t *testing.T) {
			ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)
			tc.mutate(ib)
			c, err := internal.NewFakeClient(ib)
			require.NoError(t, err)

			a := NewInitializeAction()
			a.InjectLogger(log.Log)
			a.InjectClient(c)

			target, err := a.Handle(context.TODO(), ib)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot create IntegrationBuild Job")
			assert.Contains(t, err.Error(), tc.expectedErr)
			require.NotNil(t, target)
			// The build must move to the Error phase if the mandatory parameters are missing
			assert.Equal(t, v1alpha1.IntegrationBuildPhaseError, target.Status.Phase)
			// The input must not be mutated
			assert.Equal(t, v1alpha1.IntegrationBuildPhaseNone, ib.Status.Phase)

			// No Job must be created
			var job batchv1.Job
			err = c.Get(context.TODO(), types.NamespacedName{Namespace: testNamespace, Name: testName}, &job)
			require.Error(t, err)
			assert.True(t, k8serrors.IsNotFound(err))
		})
	}
}

func TestVerifyMandatoryParams(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)
	require.NoError(t, verifyMandatoryParams(ib))

	for _, tc := range missingMandatoryParamsTests {
		t.Run(tc.name, func(t *testing.T) {
			ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)
			tc.mutate(ib)

			err := verifyMandatoryParams(ib)

			require.Error(t, err)
			assert.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestToJobExportRuntimeProviderAndGAV(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)

	job := toJob(ib)

	require.Len(t, job.Spec.Template.Spec.InitContainers, 1)
	args := job.Spec.Template.Spec.InitContainers[0].Args
	assert.Contains(t, args, "--runtime")
	assert.Equal(t, "quarkus", args[slices.Index(args, "--runtime")+1])
	assert.Contains(t, args, "--gav=org.apache.camel.integrations:my-it:1.0.0")
}

func TestToJob(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)
	ib.UID = types.UID("1234-abcd")

	job := toJob(ib)

	assert.Equal(t, testName, job.Name)
	assert.Equal(t, testNamespace, job.Namespace)
	assert.Equal(t, testName, job.Labels[camelv1.IntegrationLabel])
	assert.Equal(t, testName, job.Spec.Template.Labels[camelv1.IntegrationLabel])
	assert.Equal(t, corev1.RestartPolicyNever, job.Spec.Template.Spec.RestartPolicy)
	assert.NotEmpty(t, job.Spec.Template.Spec.Containers)

	require.Len(t, job.OwnerReferences, 1)
	owner := job.OwnerReferences[0]
	assert.Equal(t, v1alpha1.SchemeGroupVersion.String(), owner.APIVersion)
	assert.Equal(t, v1alpha1.IntegrationBuildKind, owner.Kind)
	assert.Equal(t, testName, owner.Name)
	assert.Equal(t, ib.UID, owner.UID)
	require.NotNil(t, owner.Controller)
	assert.True(t, *owner.Controller)
}
