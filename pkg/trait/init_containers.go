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

package trait

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	serving "knative.dev/serving/pkg/apis/serving/v1"

	traitv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1/trait"
)

const (
	initContainerTraitID    = "init-containers"
	initContainerTraitOrder = 1610
)

type initContainerTask struct {
	Name    string
	Image   string
	Command string
}

type initContainersTrait struct {
	BasePlatformTrait
	traitv1.InitContainersTrait `property:",squash"`

	tasks []initContainerTask
}

func newInitContainersTrait() Trait {
	return &initContainersTrait{
		BasePlatformTrait: NewBasePlatformTrait(initContainerTraitID, initContainerTraitOrder),
	}
}

func (t *initContainersTrait) Configure(e *Environment) (bool, *TraitCondition, error) {
	if e.Integration == nil || !e.IntegrationInRunningPhases() {
		return false, nil, nil
	}

	return t.parseTasks()
}

func (t *initContainersTrait) Apply(e *Environment) error {
	var initContainers *[]corev1.Container

	if err := e.Resources.VisitDeploymentE(func(deployment *appsv1.Deployment) error {
		// Deployment
		initContainers = &deployment.Spec.Template.Spec.InitContainers
		return nil
	}); err != nil {
		return err
	} else if err := e.Resources.VisitKnativeServiceE(func(service *serving.Service) error {
		// Knative Service
		initContainers = &service.Spec.ConfigurationSpec.Template.Spec.InitContainers
		return nil
	}); err != nil {
		return err
	} else if err := e.Resources.VisitCronJobE(func(cron *batchv1.CronJob) error {
		// CronJob
		initContainers = &cron.Spec.JobTemplate.Spec.Template.Spec.InitContainers
		return nil
	}); err != nil {
		return err
	}

	t.configureContainers(initContainers)

	return nil
}

func (t *initContainersTrait) configureContainers(containers *[]corev1.Container) {
	if containers == nil {
		containers = &[]corev1.Container{}
	}
	for _, task := range t.tasks {
		*containers = append(*containers, corev1.Container{
			Name:    task.Name,
			Image:   task.Image,
			Command: splitContainerCommand(task.Command),
		})
	}
}

func (t *initContainersTrait) parseTasks() (bool, *TraitCondition, error) {
	if t.Tasks == nil {
		return false, nil, nil
	}
	t.tasks = make([]initContainerTask, len(t.Tasks))
	for i, task := range t.Tasks {
		split := strings.Split(task, ";")
		if len(split) != 3 {
			return false, nil, fmt.Errorf(`could not parse init container task "%s": format expected "name;container-image;command"`, task)
		}
		t.tasks[i] = initContainerTask{
			Name:    split[0],
			Image:   split[1],
			Command: split[2],
		}
	}

	return true, nil, nil
}
