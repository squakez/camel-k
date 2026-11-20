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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/apache/camel-k/v2/pkg/apis/camel/v1alpha1"
	"github.com/apache/camel-k/v2/pkg/client"
	"github.com/apache/camel-k/v2/pkg/internal"
)

const (
	testNamespace = "default"
	testName      = "test-ib"
)

var testRequest = reconcile.Request{
	NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: testName},
}

func newTestIntegrationBuild(phase v1alpha1.IntegrationBuildPhase) *v1alpha1.IntegrationBuild {
	return &v1alpha1.IntegrationBuild{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.IntegrationBuildKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      testName,
		},
		Spec: v1alpha1.IntegrationBuildSpec{
			ExportTask: &v1alpha1.CamelJBangExporterTaskConf{
				RuntimeProvider: ptr.To("quarkus"),
				GAV:             ptr.To("org.apache.camel.integrations:my-it:1.0.0"),
			},
			BuildTask: v1alpha1.MavenJIBBuilderTaskConf{
				ImageBaseSource:         "eclipse-temurin:21-jdk",
				ImageDestination:        "registry.example.com/ns/app",
				ImageDestinationVersion: "latest",
			},
		},
		Status: v1alpha1.IntegrationBuildStatus{
			Phase: phase,
		},
	}
}

func newTestPod(phase corev1.PodPhase, terminationMessage string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      testName + "-pod",
			Labels:    map[string]string{batchv1.JobNameLabel: testName},
		},
		Status: corev1.PodStatus{Phase: phase},
	}
	if terminationMessage != "" {
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{
			{
				State: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{Message: terminationMessage},
				},
			},
		}
	}

	return pod
}

func newTestReconciler(c client.Client) *ReconcileIntegrationBuild {
	return &ReconcileIntegrationBuild{
		client:   c,
		scheme:   c.GetScheme(),
		recorder: &internal.FakeRecorder{},
		actions: []Action{
			NewInitializeAction(),
			NewMonitorAction(),
		},
	}
}

func getIntegrationBuild(t *testing.T, c client.Client) *v1alpha1.IntegrationBuild {
	t.Helper()

	var ib v1alpha1.IntegrationBuild
	require.NoError(t, c.Get(context.TODO(), testRequest.NamespacedName, &ib))

	return &ib
}

func TestReconcileIntegrationBuildNotFound(t *testing.T) {
	c, err := internal.NewFakeClient()
	require.NoError(t, err)

	result, err := newTestReconciler(c).Reconcile(context.TODO(), testRequest)

	require.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)
}

func TestReconcileIntegrationBuildInitialize(t *testing.T) {
	c, err := internal.NewFakeClient(newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone))
	require.NoError(t, err)

	result, err := newTestReconciler(c).Reconcile(context.TODO(), testRequest)

	require.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)

	// The Job is created
	var job batchv1.Job
	require.NoError(t, c.Get(context.TODO(), testRequest.NamespacedName, &job))
	require.Len(t, job.OwnerReferences, 1)
	assert.Equal(t, v1alpha1.IntegrationBuildKind, job.OwnerReferences[0].Kind)
	assert.Equal(t, testName, job.OwnerReferences[0].Name)

	// The phase moves to Building
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseBuilding, getIntegrationBuild(t, c).Status.Phase)
}

func TestReconcileIntegrationBuildMonitorNoPod(t *testing.T) {
	c, err := internal.NewFakeClient(newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding))
	require.NoError(t, err)

	result, err := newTestReconciler(c).Reconcile(context.TODO(), testRequest)

	require.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseBuilding, getIntegrationBuild(t, c).Status.Phase)
}

func TestReconcileIntegrationBuildMonitorPodRunning(t *testing.T) {
	c, err := internal.NewFakeClient(
		newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding),
		newTestPod(corev1.PodRunning, ""),
	)
	require.NoError(t, err)

	_, err = newTestReconciler(c).Reconcile(context.TODO(), testRequest)

	require.NoError(t, err)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseBuilding, getIntegrationBuild(t, c).Status.Phase)
}

func TestReconcileIntegrationBuildMonitorPodSucceeded(t *testing.T) {
	c, err := internal.NewFakeClient(
		newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding),
		newTestPod(corev1.PodSucceeded, "sha@1234567"),
	)
	require.NoError(t, err)

	_, err = newTestReconciler(c).Reconcile(context.TODO(), testRequest)

	require.NoError(t, err)
	ib := getIntegrationBuild(t, c)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseSucceeded, ib.Status.Phase)
	require.NotNil(t, ib.Status.ImageSha)
	assert.Equal(t, "sha@1234567", *ib.Status.ImageSha)
}

func TestReconcileIntegrationBuildMonitorPodFailed(t *testing.T) {
	c, err := internal.NewFakeClient(
		newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseBuilding),
		newTestPod(corev1.PodFailed, ""),
	)
	require.NoError(t, err)

	_, err = newTestReconciler(c).Reconcile(context.TODO(), testRequest)

	require.NoError(t, err)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseFailed, getIntegrationBuild(t, c).Status.Phase)
}

func TestReconcileIntegrationBuildTerminalPhaseIgnored(t *testing.T) {
	c, err := internal.NewFakeClient(newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseSucceeded))
	require.NoError(t, err)

	result, err := newTestReconciler(c).Reconcile(context.TODO(), testRequest)

	require.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseSucceeded, getIntegrationBuild(t, c).Status.Phase)
}

func TestReconcileIntegrationBuildGetError(t *testing.T) {
	c, err := internal.NewFakeClient()
	require.NoError(t, err)

	getErr := errors.New("get failed")
	//nolint:forcetypeassert
	c.(*internal.FakeClient).Intercept(&interceptor.Funcs{
		Get: func(ctx context.Context, cl ctrl.WithWatch, key ctrl.ObjectKey, obj ctrl.Object, opts ...ctrl.GetOption) error {
			return getErr
		},
	})

	_, err = newTestReconciler(c).Reconcile(context.TODO(), testRequest)

	require.ErrorIs(t, err, getErr)
}

func TestReconcileIntegrationBuildActionError(t *testing.T) {
	c, err := internal.NewFakeClient(newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone))
	require.NoError(t, err)

	createErr := errors.New("create failed")
	//nolint:forcetypeassert
	c.(*internal.FakeClient).Intercept(&interceptor.Funcs{
		Create: func(ctx context.Context, cl ctrl.WithWatch, obj ctrl.Object, opts ...ctrl.CreateOption) error {
			return createErr
		},
	})
	recorder := &internal.FakeRecorder{}
	r := newTestReconciler(c)
	r.recorder = recorder

	_, err = r.Reconcile(context.TODO(), testRequest)

	require.ErrorIs(t, err, createErr)
	assert.True(t, recorder.Called)
	assert.Equal(t, corev1.EventTypeWarning, recorder.Eventtype)
	assert.Contains(t, recorder.Message, "create failed")
	// The build must move to the Error phase
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseError, getIntegrationBuild(t, c).Status.Phase)
}

func TestReconcileIntegrationBuildMissingMandatoryParams(t *testing.T) {
	ib := newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone)
	ib.Spec.BuildTask = v1alpha1.MavenJIBBuilderTaskConf{}
	c, err := internal.NewFakeClient(ib)
	require.NoError(t, err)

	recorder := &internal.FakeRecorder{}
	r := newTestReconciler(c)
	r.recorder = recorder

	_, err = r.Reconcile(context.TODO(), testRequest)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "imageBaseSource, imageDestination or imageDestinationVersion missing")
	assert.True(t, recorder.Called)
	assert.Equal(t, corev1.EventTypeWarning, recorder.Eventtype)
	assert.Equal(t, v1alpha1.IntegrationBuildPhaseError, getIntegrationBuild(t, c).Status.Phase)

	// No Job must be created
	var jobs batchv1.JobList
	require.NoError(t, c.List(context.TODO(), &jobs))
	assert.Empty(t, jobs.Items)
}

func TestReconcileIntegrationBuildStatusPatchError(t *testing.T) {
	c, err := internal.NewFakeClient(newTestIntegrationBuild(v1alpha1.IntegrationBuildPhaseNone))
	require.NoError(t, err)

	patchErr := errors.New("patch failed")
	//nolint:forcetypeassert
	c.(*internal.FakeClient).Intercept(&interceptor.Funcs{
		Patch: func(ctx context.Context, cl ctrl.WithWatch, obj ctrl.Object, patch ctrl.Patch, opts ...ctrl.PatchOption) error {
			return patchErr
		},
		Update: func(ctx context.Context, cl ctrl.WithWatch, obj ctrl.Object, opts ...ctrl.UpdateOption) error {
			if _, ok := obj.(*v1alpha1.IntegrationBuild); ok {
				return patchErr
			}

			return cl.Update(ctx, obj, opts...)
		},
	})
	recorder := &internal.FakeRecorder{}
	r := newTestReconciler(c)
	r.recorder = recorder

	_, err = r.Reconcile(context.TODO(), testRequest)

	require.ErrorIs(t, err, patchErr)
	assert.True(t, recorder.Called)
}
