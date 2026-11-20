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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/apache/camel-k/v2/pkg/apis/camel/v1alpha1"
	"github.com/apache/camel-k/v2/pkg/client"
	"github.com/apache/camel-k/v2/pkg/internal"
	"github.com/apache/camel-k/v2/pkg/util/log"
)

func newTestMonitorAction(c client.Client) Action {
	a := NewMonitorAction()
	a.InjectLogger(log.Log)
	a.InjectClient(c)

	return a
}

func TestMonitorActionName(t *testing.T) {
	assert.Equal(t, "monitor", NewMonitorAction().Name())
}

func TestMonitorActionCanHandle(t *testing.T) {
	tests := []struct {
		phase    v1alpha1.IntegrationBuildPhase
		expected bool
	}{
		{v1alpha1.IntegrationBuildPhaseNone, false},
		{v1alpha1.IntegrationBuildPhaseBuilding, true},
		{v1alpha1.IntegrationBuildPhaseSucceeded, false},
		{v1alpha1.IntegrationBuildPhaseFailed, false},
	}

	a := NewMonitorAction()
	for _, tc := range tests {
		t.Run(string(tc.phase), func(t *testing.T) {
			assert.Equal(t, tc.expected, a.CanHandle(newTestIntegrationBuild(tc.phase)))
		})
	}
}

func TestMonitorActionHandleNoPod(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding)
	c, err := internal.NewFakeClient(ib)
	require.NoError(t, err)

	target, err := newTestMonitorAction(c).Handle(context.TODO(), ib)

	// nil target means "nothing to update"
	require.NoError(t, err)
	assert.Nil(t, target)
}

func TestMonitorActionHandlePodPending(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding)
	c, err := internal.NewFakeClient(ib, newTestPod(corev1.PodPending, ""))
	require.NoError(t, err)

	target, err := newTestMonitorAction(c).Handle(context.TODO(), ib)

	require.NoError(t, err)
	require.NotNil(t, target)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseBuilding, target.Status.Phase)
}

func TestMonitorActionHandlePodSucceeded(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding)
	c, err := internal.NewFakeClient(ib, newTestPod(corev1.PodSucceeded, "sha@abc"))
	require.NoError(t, err)

	target, err := newTestMonitorAction(c).Handle(context.TODO(), ib)

	require.NoError(t, err)
	require.NotNil(t, target)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseSucceeded, target.Status.Phase)
	require.NotNil(t, target.Status.ImageSha)
	assert.Equal(t, "sha@abc", *target.Status.ImageSha)
	// The input must not be mutated
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseBuilding, ib.Status.Phase)
}

func TestMonitorActionHandlePodSucceededNoTerminationMessage(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding)
	c, err := internal.NewFakeClient(ib, newTestPod(corev1.PodSucceeded, ""))
	require.NoError(t, err)

	target, err := newTestMonitorAction(c).Handle(context.TODO(), ib)

	require.NoError(t, err)
	require.NotNil(t, target)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseSucceeded, target.Status.Phase)
	require.NotNil(t, target.Status.ImageSha)
	assert.Empty(t, *target.Status.ImageSha)
	// The image version is not populated yet
	require.NotNil(t, target.Status.ImageVersion)
	assert.Empty(t, *target.Status.ImageVersion)
}

func TestMonitorActionHandlePodFailed(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding)
	c, err := internal.NewFakeClient(ib, newTestPod(corev1.PodFailed, ""))
	require.NoError(t, err)

	target, err := newTestMonitorAction(c).Handle(context.TODO(), ib)

	require.NoError(t, err)
	require.NotNil(t, target)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseFailed, target.Status.Phase)
	assert.Nil(t, target.Status.ImageSha)
}

func TestMonitorActionHandleListError(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding)
	c, err := internal.NewFakeClient(ib)
	require.NoError(t, err)

	listErr := errors.New("list failed")
	//nolint:forcetypeassert
	c.(*internal.FakeClient).Intercept(&interceptor.Funcs{
		List: func(ctx context.Context, cl ctrl.WithWatch, list ctrl.ObjectList, opts ...ctrl.ListOption) error {
			return listErr
		},
	})

	target, err := newTestMonitorAction(c).Handle(context.TODO(), ib)

	require.ErrorIs(t, err, listErr)
	assert.NotNil(t, target)
}

func TestMonitorActionHandleMostRecentPodWins(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding)

	now := time.Now()
	oldPod := newTestPod(corev1.PodFailed, "")
	oldPod.Name = "old"
	oldPod.CreationTimestamp = metav1.NewTime(now.Add(-time.Hour))
	newPod := newTestPod(corev1.PodSucceeded, "sha@new")
	newPod.Name = "new"
	newPod.CreationTimestamp = metav1.NewTime(now)

	c, err := internal.NewFakeClient(ib, oldPod, newPod)
	require.NoError(t, err)

	target, err := newTestMonitorAction(c).Handle(context.TODO(), ib)

	require.NoError(t, err)
	require.NotNil(t, target)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseSucceeded, target.Status.Phase)
	require.NotNil(t, target.Status.ImageSha)
	assert.Equal(t, "sha@new", *target.Status.ImageSha)
}

func TestGetPodBuildIgnoresUnrelatedPods(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding)

	otherNamespace := newTestPod(corev1.PodSucceeded, "sha@other")
	otherNamespace.Name = "other-namespace"
	otherNamespace.Namespace = "another"

	noLabel := newTestPod(corev1.PodSucceeded, "sha@nolabel")
	noLabel.Name = "no-label"
	noLabel.Labels = nil

	otherJob := newTestPod(corev1.PodSucceeded, "sha@otherjob")
	otherJob.Name = "other-job"
	otherJob.Labels = map[string]string{batchv1.JobNameLabel: "another-job"}

	c, err := internal.NewFakeClient(ib, otherNamespace, noLabel, otherJob)
	require.NoError(t, err)

	action, ok := NewMonitorAction().(*monitorAction)
	require.True(t, ok)
	action.InjectClient(c)

	pod, err := action.getPodBuild(context.TODO(), ib)

	require.NoError(t, err)
	assert.Nil(t, pod)
}

func TestGetBuildImageInfo(t *testing.T) {
	terminated := func(msg string) corev1.ContainerStatus {
		return corev1.ContainerStatus{
			State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Message: msg}},
		}
	}
	running := corev1.ContainerStatus{
		State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
	}

	tests := []struct {
		name     string
		statuses []corev1.ContainerStatus
		expected string
	}{
		{"no container statuses", nil, ""},
		{"no terminated container", []corev1.ContainerStatus{running}, ""},
		{"single terminated container", []corev1.ContainerStatus{terminated("sha@1")}, "sha@1"},
		{"last terminated container wins", []corev1.ContainerStatus{terminated("sha@1"), terminated("sha@2")}, "sha@2"},
		{"running container does not override", []corev1.ContainerStatus{terminated("sha@1"), running}, "sha@1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pod := &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: tc.statuses}}

			sha, version := getBuildImageInfo(pod)

			assert.Equal(t, tc.expected, sha)
			assert.Empty(t, version)
		})
	}
}
