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
	"strings"

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

	job, err := toJob(ib)
	if err != nil {
		return target, fmt.Errorf("cannot transform IntegrationBuild to a Job: %w", err)
	}
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
	defaultExportImage   = "apache/camel-jbang:4.22.1"
	defaultGitCloneImage = "alpine/git"
	defaultBuildImage    = "maven:3.9.11-eclipse-temurin-21"

	workspaceVolume      = "workspace"
	workspaceDir         = "/workspace"
	inputVolume          = "input"
	inputDir             = "/input"
	profileVolume        = "profile"
	profileDir           = "/profile"
	profileConfigmapName = "camel-git-profile"

	profileScript = `set -e
echo "Cloning repository..."
git clone --depth=1 https://github.com/squakez/sample.git /workspace
echo "Injecting Jib profile..."
sh /profile/jib-profiler.sh /workspace/pom.xml /profile/jib-profile.xml
`
)

// toJob converts the IntegrationBuild into a Job running the build pipeline.
func toJob(ib *v1alpha1.IntegrationBuild) (*batchv1.Job, error) {
	var securityContext *v1.PodSecurityContext
	if ib.Spec.SecurityContextRunAsUser != nil {
		securityContext = &v1.PodSecurityContext{
			RunAsUser: ptr.To(int64(*ib.Spec.SecurityContextRunAsUser)),
		}
	}

	generateProjectContainer, err := exportOrClone(ib)
	if err != nil {
		return nil, err
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
					InitContainers:  []v1.Container{*generateProjectContainer},
					Containers:      []v1.Container{*buildContainer(&ib.Spec.BuildTask)},
					Volumes:         volumes(ib),
				},
			},
		},
	}, nil
}

func workspaceEnv() []v1.EnvVar {
	return []v1.EnvVar{{Name: "HOME", Value: workspaceDir}}
}

// Return either a Camel JBang export container task or a Git clone task. If none or more than a single is provided,
// then, it returns an error
func exportOrClone(ib *v1alpha1.IntegrationBuild) (*v1.Container, error) {
	if ib.Spec.ExportTask == nil && ib.Spec.GitCloneTask == nil {
		return nil, errors.New("no export or clone tasks configured. You need exactly one.")
	}
	if ib.Spec.ExportTask != nil && ib.Spec.GitCloneTask != nil {
		return nil, errors.New("both export and clone tasks configured. You need exactly one.")
	}

	if ib.Spec.GitCloneTask != nil {
		return gitCloneContainer(ib.Spec.GitCloneTask), nil
	}

	// no other possible combinations
	return exportContainer(ib.Spec.ExportTask), nil
}

// creates a Camel Export container task from configuration.
func exportContainer(conf *v1alpha1.CamelJBangExporterTaskConf) *v1.Container {
	image := defaultExportImage
	if conf.ImageTask != nil {
		image = *conf.ImageTask
	}
	commands := []string{"/bin/sh", "-c"}

	args := []string{"camel", "export", inputDir, "--dir", workspaceDir}
	if conf.RuntimeProvider != nil {
		args = append(args, "--runtime", *conf.RuntimeProvider)
	}
	if conf.GAV != nil {
		args = append(args, "--gav="+*conf.GAV)
	}

	return &v1.Container{
		Name:       "camel-export",
		Image:      image,
		WorkingDir: workspaceDir,
		Env:        workspaceEnv(),
		Command:    commands,
		// We need to make it a single command as it runs in shell
		Args: []string{strings.Join(args, " ")},
		VolumeMounts: []v1.VolumeMount{
			{Name: workspaceVolume, MountPath: workspaceDir},
			{Name: inputVolume, MountPath: inputDir},
		},
	}
}

// creates a Git Clone container task from configuration.
func gitCloneContainer(conf *v1alpha1.GitCloneTaskConf) *v1.Container {
	image := defaultGitCloneImage
	commands := []string{"/bin/sh", "-c"}
	// TODO: add git parameters configuration
	args := []string{profileScript}

	return &v1.Container{
		Name:       "git-clone",
		Image:      image,
		WorkingDir: workspaceDir,
		Env:        workspaceEnv(),
		Command:    commands,
		Args:       args,
		VolumeMounts: []v1.VolumeMount{
			{Name: workspaceVolume, MountPath: workspaceDir},
			{Name: profileVolume, MountPath: profileDir},
		},
	}
}

// creates a Maven build container task from configuration.
func buildContainer(conf *v1alpha1.MavenJIBBuilderTaskConf) *v1.Container {
	image := defaultBuildImage
	if conf.ImageTask != nil {
		image = *conf.ImageTask
	}

	args := []string{
		"package",
		"jib:build",
		"-Pjib",
		"-DskipTests",
		"--no-transfer-progress",
		fmt.Sprintf("-Djib.to.image=%s:%s", conf.ImageDestination, conf.ImageDestinationVersion),
		"-Djib.from.image=" + conf.ImageBaseSource,
	}
	if conf.ImageContainerUser != nil {
		args = append(args, fmt.Sprintf("-Djib.container.user=%d", *conf.ImageContainerUser))
	}
	args = append(args, "-Djib.disableUpdateChecks=true")
	if conf.AllowInsecureRegistries != nil {
		args = append(args, fmt.Sprintf("-Djib.allowInsecureRegistries=%t", *conf.AllowInsecureRegistries))
	}
	args = append(args, "-Djib.outputPaths.digest=/dev/termination-log")

	return &v1.Container{
		Name:       "maven-build",
		Image:      image,
		WorkingDir: workspaceDir,
		Command:    []string{"mvn"},
		Args:       args,
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
	volumes := []v1.Volume{
		{
			Name:         workspaceVolume,
			VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
		},
	}

	// Only add sources if the project is going to be exported from sources
	if ib.Spec.ExportTask != nil {
		sources := make([]v1.VolumeProjection, 0, len(ib.Spec.ExportTask.Sources))
		for _, name := range ib.Spec.ExportTask.Sources {
			sources = append(sources, v1.VolumeProjection{
				ConfigMap: &v1.ConfigMapProjection{
					LocalObjectReference: v1.LocalObjectReference{Name: name},
				},
			})
		}

		volumes = append(volumes, v1.Volume{
			Name: inputVolume,
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{Sources: sources},
			},
		})
	}

	// Only add profile if the project is going to be cloned from Git
	// TODO: create a configmap with the profiler if it does not exist. Also, for security reasons
	// verify via hashing that the content is the one we expect, in order to avoid possible tampering.
	if ib.Spec.GitCloneTask != nil {
		volumes = append(volumes, v1.Volume{
			Name: profileVolume,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: profileConfigmapName,
					},
				},
			},
		})
	}

	return volumes
}
