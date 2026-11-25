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
	"sort"

	"github.com/apache/camel-k/v2/pkg/apis/camel/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

// NewMonitorAction returns an action that monitors the IntegrationBuild pipeline.
func NewMonitorAction() Action {
	return &monitorAction{}
}

type monitorAction struct {
	baseAction
}

func (action *monitorAction) Name() string {
	return "monitor"
}

func (action *monitorAction) CanHandle(ib *v1alpha1.IntegrationBuild) bool {
	return ib.Status.Phase == v1alpha1.IntegrationBuildPhaseBuilding
}

func (action *monitorAction) Handle(ctx context.Context, ib *v1alpha1.IntegrationBuild) (*v1alpha1.IntegrationBuild, error) {
	action.L.Info("Monitoring IntegrationBuild")
	target := ib.DeepCopy()

	podBuild, err := action.getPodBuild(ctx, ib)
	if err != nil {
		return target, err
	}
	if podBuild == nil {
		return nil, nil
	}

	status := v1alpha1.IntegrationBuildStatus{}

	if podBuild.Status.Phase == corev1.PodSucceeded {
		imageSha, imageVersion := getBuildImageInfo(podBuild)
		status.Phase = v1alpha1.IntegrationBuildPhaseSucceeded
		status.ImageSha = ptr.To(imageSha)
		status.ImageVersion = ptr.To(imageVersion)
	} else if podBuild.Status.Phase == corev1.PodFailed {
		status.Phase = v1alpha1.IntegrationBuildPhaseFailed
		// TODO: add a reason why it failed
	} else {
		// Pod still running or pending
		return target, nil
	}

	target.Status = status

	return target, nil
}

func (action *monitorAction) getPodBuild(ctx context.Context, ib *v1alpha1.IntegrationBuild) (*corev1.Pod, error) {
	podList := &corev1.PodList{}
	err := action.client.List(ctx, podList,
		ctrl.InNamespace(ib.Namespace),
		ctrl.MatchingLabels{
			batchv1.JobNameLabel: ib.Name,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return nil, nil
	}

	// In case of multiple pods (e.g. retries), sort to get the most recent one
	sort.Slice(podList.Items, func(i, j int) bool {
		return podList.Items[j].CreationTimestamp.Before(&podList.Items[i].CreationTimestamp)
	})

	return &podList.Items[0], nil
}

// Scan the pod container statuses and retrieve the SHA of the container image produced.
func getBuildImageInfo(pod *corev1.Pod) (string, string) {
	imageSha := ""
	imageVersion := ""
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Terminated != nil {
			imageSha = containerStatus.State.Terminated.Message
		}
	}
	return imageSha, imageVersion
}
