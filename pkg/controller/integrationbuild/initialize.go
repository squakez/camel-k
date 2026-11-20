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
	"fmt"

	camelv1 "github.com/apache/camel-k/v2/pkg/apis/camel/v1"
	"github.com/apache/camel-k/v2/pkg/apis/camel/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// NewInitializeAction returns a action that initializes the IntegrationBuild pipeline.
func NewInitializeAction() Action {
	return &initializeAction{}
}

type initializeAction struct {
	baseAction
}

func (action *initializeAction) Name() string {
	return "initialize"
}

func (action *initializeAction) CanHandle(ib *v1alpha1.IntegrationBuild) bool {
	return ib.Status.Phase == v1alpha1.IntegrationBuildPhaseNone
}

func (action *initializeAction) Handle(ctx context.Context, ib *v1alpha1.IntegrationBuild) (*v1alpha1.IntegrationBuild, error) {
	action.L.Info("Initializing IntegrationBuild")
	target := ib.DeepCopy()

	if err := verifyMandatoryParams(ib); err != nil {
		target.Status = v1alpha1.IntegrationBuildStatus{
			Phase: v1alpha1.IntegrationBuildPhaseError,
		}
		return target, fmt.Errorf("cannot create IntegrationBuild Job: %w", err)
	}

	job := toJob(ib)
	if err := action.client.Create(ctx, job); err != nil {
		target.Status = v1alpha1.IntegrationBuildStatus{
			Phase: v1alpha1.IntegrationBuildPhaseError,
		}
		return target, fmt.Errorf("cannot create IntegrationBuild Job: %w", err)
	}

	target.Status = v1alpha1.IntegrationBuildStatus{
		Phase: v1alpha1.IntegrationBuildPhaseBuilding,
	}

	return target, nil
}

// TODO, parametrize
const (
	defaultExportImage = "apache/camel-jbang:4.21.0"
	defaultBuildImage  = "maven:3.9.11-eclipse-temurin-21"

	workspaceVolume = "workspace"
	workspaceDir    = "/workspace"
	inputVolume     = "input"
	inputDir        = "/input"
)

// Verify that all mandatory parameters are there
func verifyMandatoryParams(ib *v1alpha1.IntegrationBuild) error {
	if ib.Spec.BuildTask.ImageBaseSource == "" || ib.Spec.BuildTask.ImageDestination == "" || ib.Spec.BuildTask.ImageDestinationVersion == "" {
		return errors.New("imageBaseSource, imageDestination or imageDestinationVersion missing")
	}
	task := ib.Spec.ExportTask
	if task == nil || task.RuntimeProvider == nil || *task.RuntimeProvider == "" || task.GAV == nil || *task.GAV == "" {
		return errors.New("runtimeProvider or gav missing")
	}

	return nil
}

// toJob converts the IntegrationBuild into a Job running the build pipeline.
func toJob(ib *v1alpha1.IntegrationBuild) *batchv1.Job {
	var securityContext *v1.PodSecurityContext
	if ib.Spec.SecurityContextRunAsUser != nil {
		securityContext = &v1.PodSecurityContext{
			RunAsUser: ptr.To(int64(*ib.Spec.SecurityContextRunAsUser)),
		}
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ib.Name,
			Namespace: ib.Namespace,
			Labels: map[string]string{
				// Required to cache the job
				// TODO: extract the Integration name from ib owner
				camelv1.IntegrationLabel: ib.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1alpha1.SchemeGroupVersion.String(),
					Kind:       v1alpha1.IntegrationBuildKind,
					Name:       ib.Name,
					UID:        ib.UID,
					Controller: ptr.To(true),
				},
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: ptr.To(int32(0)),
			Template: v1.PodTemplateSpec{
				Labels: map[string]string{
					// Required to cache the Pod
					// TODO: extract the Integration name from ib owner
					camelv1.IntegrationLabel: ib.Name,
				},
				Spec: v1.PodSpec{
					SecurityContext: securityContext,
					RestartPolicy:   v1.RestartPolicyNever,
					InitContainers:  []v1.Container{exportContainer(ib)},
					Containers:      []v1.Container{buildContainer(ib)},
					Volumes:         volumes(ib),
				},
			},
		},
	}
}

func workspaceEnv() []v1.EnvVar {
	return []v1.EnvVar{{Name: "HOME", Value: workspaceDir}}
}

func exportContainer(ib *v1alpha1.IntegrationBuild) v1.Container {
	image := defaultExportImage
	args := []string{"export"}
	// TODO: should we filter or can we use them all?
	args = append(args, "/input/*")
	args = append(args, "--dir", workspaceDir)
	if task := ib.Spec.ExportTask; task != nil {
		if task.ImageTask != nil {
			image = *task.ImageTask
		}
		if task.RuntimeProvider != nil {
			args = append(args, "--runtime", *task.RuntimeProvider)
		}
		if task.GAV != nil {
			args = append(args, "--gav="+*task.GAV)
		}
	}

	return v1.Container{
		Name:       "camel-export",
		Image:      image,
		WorkingDir: workspaceDir,
		Env:        workspaceEnv(),
		Args:       args,
		VolumeMounts: []v1.VolumeMount{
			{Name: workspaceVolume, MountPath: workspaceDir},
			{Name: inputVolume, MountPath: inputDir},
		},
	}
}

func buildContainer(ib *v1alpha1.IntegrationBuild) v1.Container {
	task := ib.Spec.BuildTask
	image := defaultBuildImage
	if task.ImageTask != nil {
		image = *task.ImageTask
	}

	args := []string{
		"package",
		"jib:build",
		"-Pjib",
		"-f",
		workspaceDir + "/pom.xml",
		"-DskipTests",
		"--no-transfer-progress",
		fmt.Sprintf("-Djib.to.image=%s:%s", task.ImageDestination, task.ImageDestinationVersion),
		"-Djib.from.image=" + task.ImageBaseSource,
	}
	if task.ImageContainerUser != nil {
		args = append(args, fmt.Sprintf("-Djib.container.user=%d", *task.ImageContainerUser))
	}
	args = append(args, "-Djib.disableUpdateChecks=true")
	if task.AllowInsecureRegistries != nil {
		args = append(args, fmt.Sprintf("-Djib.allowInsecureRegistries=%t", *task.AllowInsecureRegistries))
	}
	args = append(args, "-Djib.outputPaths.digest=/dev/termination-log")

	return v1.Container{
		Name:    "maven-build",
		Image:   image,
		Command: []string{"mvn"},
		Args:    args,
		Env: append(workspaceEnv(), v1.EnvVar{
			Name:  "MAVEN_CONFIG",
			Value: workspaceDir + "/.m2",
		}),
		VolumeMounts: []v1.VolumeMount{
			{Name: workspaceVolume, MountPath: workspaceDir},
		},
	}
}

func volumes(ib *v1alpha1.IntegrationBuild) []v1.Volume {
	sources := make([]v1.VolumeProjection, 0, len(ib.Spec.Sources))
	for _, name := range ib.Spec.Sources {
		sources = append(sources, v1.VolumeProjection{
			ConfigMap: &v1.ConfigMapProjection{
				LocalObjectReference: v1.LocalObjectReference{Name: name},
			},
		})
	}

	return []v1.Volume{
		{
			Name:         workspaceVolume,
			VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
		},
		{
			Name: inputVolume,
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{Sources: sources},
			},
		},
	}
}
