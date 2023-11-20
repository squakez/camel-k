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
	"path/filepath"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	serving "knative.dev/serving/pkg/apis/serving/v1"

	v1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	traitv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1/trait"
	"github.com/apache/camel-k/v2/pkg/util/kubernetes"
	"github.com/apache/camel-k/v2/pkg/util/resource"
	utilResource "github.com/apache/camel-k/v2/pkg/util/resource"
)

type mountTrait struct {
	BasePlatformTrait
	traitv1.MountTrait `property:",squash"`
}

func newMountTrait() Trait {
	return &mountTrait{
		// Must follow immediately the container trait
		BasePlatformTrait: NewBasePlatformTrait("mount", 1610),
	}
}

func (t *mountTrait) Configure(e *Environment) (bool, *TraitCondition, error) {
	if e.Integration == nil || !e.IntegrationInRunningPhases() {
		return false, nil, nil
	}

	// Validate resources and pvcs
	for _, c := range t.Configs {
		if !strings.HasPrefix(c, "configmap:") && !strings.HasPrefix(c, "secret:") {
			return false, nil, fmt.Errorf("unsupported config %s, must be a configmap or secret resource", c)
		}
	}
	for _, r := range t.Resources {
		if !strings.HasPrefix(r, "configmap:") && !strings.HasPrefix(r, "secret:") {
			return false, nil, fmt.Errorf("unsupported resource %s, must be a configmap or secret resource", r)
		}
	}

	// mount trait needs always to be executed when there are sources
	return len(e.Integration.Sources()) > 0, nil, nil
}

func (t *mountTrait) Apply(e *Environment) error {
	container := e.GetIntegrationContainer()
	if container == nil {
		return fmt.Errorf("unable to find integration container: %s", e.Integration.Name)
	}

	var volumes *[]corev1.Volume
	visited := false

	// Deployment
	if err := e.Resources.VisitDeploymentE(func(deployment *appsv1.Deployment) error {
		volumes = &deployment.Spec.Template.Spec.Volumes
		visited = true
		return nil
	}); err != nil {
		return err
	}

	// Knative Service
	if err := e.Resources.VisitKnativeServiceE(func(service *serving.Service) error {
		volumes = &service.Spec.ConfigurationSpec.Template.Spec.Volumes
		visited = true
		return nil
	}); err != nil {
		return err
	}

	// CronJob
	if err := e.Resources.VisitCronJobE(func(cron *batchv1.CronJob) error {
		volumes = &cron.Spec.JobTemplate.Spec.Template.Spec.Volumes
		visited = true
		return nil
	}); err != nil {
		return err
	}

	if visited {
		// Volumes declared in the Integration resources
		e.configureVolumesAndMounts(volumes, &container.VolumeMounts)
		// Volumes declared in the trait config/resource options
		err := t.configureVolumesAndMounts(volumes, &container.VolumeMounts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *mountTrait) configureVolumesAndMounts(vols *[]corev1.Volume, mnts *[]corev1.VolumeMount) error {
	for _, c := range t.Configs {
		if conf, parseErr := utilResource.ParseConfig(c); parseErr == nil {
			t.mountResource(vols, mnts, conf)
		} else {
			return parseErr
		}
	}
	for _, r := range t.Resources {
		if res, parseErr := utilResource.ParseResource(r); parseErr == nil {
			t.mountResource(vols, mnts, res)
		} else {
			return parseErr
		}
	}
	for _, v := range t.Volumes {
		if vol, parseErr := utilResource.ParseVolume(v); parseErr == nil {
			t.mountResource(vols, mnts, vol)
		} else {
			return parseErr
		}
	}

	return nil
}

func (t *mountTrait) mountResource(vols *[]corev1.Volume, mnts *[]corev1.VolumeMount, conf *utilResource.Config) {
	refName := kubernetes.SanitizeLabel(conf.Name())
	dstDir := conf.DestinationPath()
	dstFile := ""
	if conf.DestinationPath() != "" {
		if conf.Key() != "" {
			dstFile = filepath.Base(conf.DestinationPath())
		} else {
			dstFile = conf.Key()
		}
	}
	vol := getVolume(refName, string(conf.StorageType()), conf.Name(), conf.Key(), dstFile)
	mntPath := getMountPoint(conf.Name(), dstDir, string(conf.StorageType()), string(conf.ContentType()))
	readOnly := true
	if conf.StorageType() == utilResource.StorageTypePVC {
		readOnly = false
	}
	mnt := getMount(refName, mntPath, dstFile, readOnly)

	if !exists(*vols, *vol) {
		*vols = append(*vols, *vol)
		*mnts = append(*mnts, *mnt)
	} else {
		t.L.Infof("Volume %s already exists in the deployment, won't include it again.", vol.Name)
	}
}

func exists(vols []corev1.Volume, vol corev1.Volume) bool {
	for _, v := range vols {
		if v.Name == vol.Name {
			return true
		}
	}
	return false
}

func (t *mountTrait) Reverse(e *Environment, traits *v1.Traits) error {
	deploymentName := e.Integration.Annotations["camel.apache.org/imported-by"]
	deploy, err := e.Client.AppsV1().Deployments(e.Integration.Namespace).Get(e.Ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if deploy.Spec.Template.Spec.Volumes == nil || len(deploy.Spec.Template.Spec.Volumes) == 0 {
		return nil
	}
	if traits.Mount == nil {
		traits.Mount = &traitv1.MountTrait{
			Configs: make([]string, 0),
		}
	}
	for _, vol := range deploy.Spec.Template.Spec.Volumes {
		var config *resource.Config
		switch {
		case vol.ConfigMap != nil:
			// TODO must parse the real value, this is just for POC purposes
			// We must also improve the trait in order to let the user specify the volume name
			if config, err = resource.ParseConfig("configmap:my-cm@/tmp/app/data"); err != nil {
				return err
			}
			traits.Mount.Configs = append(traits.Mount.Configs, config.String())
		}
	}
	return nil
}
