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
	"os"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/client/camel/clientset/versioned/scheme"
	"github.com/apache/camel-k/v2/pkg/trait"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

// NewImportAction creates a new import action.
func NewImportAction(reader ctrl.Reader) Action {
	return &importAction{
		reader: reader,
	}
}

type importAction struct {
	baseAction
	reader ctrl.Reader
}

// Name returns a common name of the action.
func (action *importAction) Name() string {
	return "initialize"
}

// CanHandle tells whether this action can handle the integration.
func (action *importAction) CanHandle(integration *v1.Integration) bool {
	return integration.Status.Phase == v1.IntegrationPhaseImporting || integration.Status.Phase == v1.IntegrationPhaseRunning
}

// Handle handles the integrations.
func (action *importAction) Handle(ctx context.Context, it *v1.Integration) (*v1.Integration, error) {
	if it.Status.Phase == v1.IntegrationPhaseRunning {
		return action.inspectPod(ctx, it)
	}
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

func (action *importAction) inspectPod(ctx context.Context, it *v1.Integration) (*v1.Integration, error) {
	pod, err := getItPod(ctx, action.reader, it)
	if err != nil {
		fmt.Println("Error while loading pod:", err)
		return it, err
	}
	if err = action.getPom(ctx, pod); err != nil {
		if err != nil {
			fmt.Println("Error while reading from pod:", err)
			return it, err
		}
	}

	return it, nil
}

func getItPod(ctx context.Context, c ctrl.Reader, it *v1.Integration) (*corev1.Pod, error) {
	pods := corev1.PodList{}
	err := c.List(ctx, &pods,
		ctrl.InNamespace(it.Namespace),
		ctrl.MatchingLabels{v1.IntegrationLabel: it.Name},
	)
	if err != nil {
		return nil, err
	}
	return &pods.Items[0], nil
}

func (action *importAction) getPom(ctx context.Context, pod *corev1.Pod) error {
	fmt.Println("Reading from Pod", pod.GetName())

	for _, container := range pod.Status.ContainerStatuses {
		fmt.Println("Executing on container", container.Name, container.State)
		if container.State.Running == nil {
			continue
		}
		r := action.client.CoreV1().RESTClient().Post().
			Resource("pods").
			Namespace(pod.Namespace).
			Name(pod.Name).
			SubResource("exec").
			Param("container", container.Name)

		r.VersionedParams(&corev1.PodExecOptions{
			Container: container.Name,
			//Command:   []string{"/bin/bash", "-c", "cd /deployments/ && jar -xvf my-camel-app.jar && cat META-INF/maven/org.acme/my-service/pom.xml"},
			Command: []string{"/bin/bash", "-c", "kill -SIGTERM 1"},
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

		exec, err := remotecommand.NewSPDYExecutor(action.client.GetConfig(), "POST", r.URL())
		if err != nil {
			fmt.Println("NewSPDYExecutor error")
			return err
		}

		err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Tty:    false,
		})
		if err != nil {
			fmt.Println("StreamWithContext error")
			return err
		}
	}

	return nil
}
